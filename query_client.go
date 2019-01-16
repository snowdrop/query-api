package main

import (
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"github.com/liggitt/tabwriter"
	codescheme "github.com/snowdrop/query-api/pkg/scheme"
	printersinternal "k8s.io/kubernetes/pkg/printers/internalversion"
	"k8s.io/client-go/rest"
	"os"
	"reflect"
)

var (
	scheme = runtime.NewScheme()
)

const (
	tabwriterMinWidth = 6
	tabwriterWidth    = 4
	tabwriterPadding  = 3
	tabwriterPadChar  = ' '
	tabwriterFlags    = tabwriter.RememberWidths
)

func main() {
	o := NewParams(genericclioptions.ConfigFlags{})
	o.configFlags.ToRESTConfig()
	o.builder = resource.NewBuilder(o.configFlags)

	o.ToPrinter = func(mapping *meta.RESTMapping, withNamespace bool, withKind bool) (printers.ResourcePrinterFunc, error) {
		// make a new copy of current flags / opts before mutating
		printFlags := o.PrintFlags.Copy()

		if mapping != nil {
			printFlags.SetKind(mapping.GroupVersionKind.GroupKind())
		}
		if withNamespace {
			printFlags.EnsureWithNamespace()
		}
		if withKind {
			printFlags.EnsureWithKind()
		}

		printer, err := printFlags.ToPrinter()
		if err != nil {
			return nil, err
		}

		return printer.PrintObj, nil
	}

	r := o.builder.
		Unstructured().
		NamespaceParam("my-spring-boot").
		ResourceTypeOrNameArgs(true, "all").
		LabelSelector("app=my-spring-boot ").
		Latest().
		Flatten().
		TransformRequests(func(req *rest.Request) {
			group := metav1beta1.GroupName
			version := metav1beta1.SchemeGroupVersion.Version

			tableParam := fmt.Sprintf("application/json;as=Table;v=%s;g=%s, application/json", version, group)
			req.SetHeader("Accept", tableParam)
		}).
		Do()

	allErrs := []error{}
	errs := sets.NewString()
	infos, err := r.Infos()
	if err != nil {
		allErrs = append(allErrs, err)
	}

	//PrintResult(infos)
	printWithKind := multipleGVKsRequested(infos)

	objs := make([]runtime.Object, len(infos))
	for ix := range infos {
		table, err := DecodeIntoTable(infos[ix].Object)
		if err == nil {
			infos[ix].Object = table
		} else {
			// if we are unable to decode server response into a v1beta1.Table,
			// fallback to client-side printing with whatever info the server returned.
			fmt.Printf("Unable to decode server response into a Table. Falling back to hardcoded types: %v", err)
		}
		objs[ix] = infos[ix].Object
	}

	var positioner OriginalPositioner
	var printer printers.ResourcePrinter
	var lastMapping *meta.RESTMapping
	nonEmptyObjCount := 0

	_, wFile, _ := os.Pipe()
	os.Stdout = wFile
	o.Out = os.Stdout

	w := GetNewTabWriter(o.Out)
	for ix := range objs {
		var mapping *meta.RESTMapping
		var info *resource.Info
		if positioner != nil {
			info = infos[positioner.OriginalPosition(ix)]
			mapping = info.Mapping
		} else {
			info = infos[ix]
			mapping = info.Mapping
		}

		// if dealing with a table that has no rows, skip remaining steps
		// and avoid printing an unnecessary newline
		if table, isTable := info.Object.(*metav1beta1.Table); isTable {
			if len(table.Rows) == 0 {
				continue
			}
		}

		nonEmptyObjCount++

		// TODO - use parameter to configure
		printWithNamespace := false

		if mapping != nil && mapping.Scope.Name() == meta.RESTScopeNameRoot {
			printWithNamespace = false
		}

		if shouldGetNewPrinterForMapping(printer, lastMapping, mapping) {
			w.Flush()
			w.SetRememberedWidths(nil)

			// TODO: this doesn't belong here
			// add linebreak between resource groups (if there is more than one)
			// skip linebreak above first resource group
			if lastMapping != nil && !o.NoHeaders {
				fmt.Fprintln(o.ErrOut)
			}

			printer, err = o.ToPrinter(mapping, printWithNamespace, printWithKind)
			if err != nil {
				if !errs.Has(err.Error()) {
					errs.Insert(err.Error())
					allErrs = append(allErrs, err)
				}
				continue
			}

			lastMapping = mapping
		}

		// ensure a versioned object is passed to the custom-columns printer
		// if we are using OpenAPI columns to print
		if o.PrintWithOpenAPICols {
			printer.PrintObj(info.Object, w)
			continue
		}

		printer.PrintObj(info.Object, w)
	}
	w.Flush()

}

// SetKind sets the Kind option
func (f *HumanPrintFlags) SetKind(kind schema.GroupKind) {
	f.Kind = kind
}

// SetKind sets the Kind option of humanreadable flags
func (f *PrintFlags) SetKind(kind schema.GroupKind) {
	f.HumanReadableFlags.SetKind(kind)
}

// EnsureWithKind sets the "Showkind" humanreadable option to true.
func (f *HumanPrintFlags) EnsureWithKind() error {
	showKind := true
	f.ShowKind = &showKind
	return nil
}

// EnsureWithKind ensures that humanreadable flags return
// a printer capable of including resource kinds.
func (f *PrintFlags) EnsureWithKind() error {
	return f.HumanReadableFlags.EnsureWithKind()
}

// Copy returns a copy of PrintFlags for mutation
func (f *PrintFlags) Copy() PrintFlags {
	printFlags := *f
	return printFlags
}

// AllowedFormats is the list of formats in which data can be displayed
func (f *PrintFlags) AllowedFormats() []string {
	formats := f.JSONYamlPrintFlags.AllowedFormats()
	formats = append(formats, f.NamePrintFlags.AllowedFormats()...)
	formats = append(formats, f.TemplateFlags.AllowedFormats()...)
	return formats
}

// EnsureWithNamespace sets the "WithNamespace" humanreadable option to true.
func (f *HumanPrintFlags) EnsureWithNamespace() error {
	f.WithNamespace = true
	return nil
}

// EnsureWithNamespace ensures that humanreadable flags return
// a printer capable of printing with a "namespace" column.
func (f *PrintFlags) EnsureWithNamespace() error {
	return f.HumanReadableFlags.EnsureWithNamespace()
}

// Copy returns a copy of PrintFlags for mutation
func (f *PrintFlags) Copy() PrintFlags {
	printFlags := *f
	return printFlags
}

// AllowedFormats returns more customized formating options
func (f *HumanPrintFlags) AllowedFormats() []string {
	return []string{"wide"}
}

// ToPrinter receives an outputFormat and returns a printer capable of
// handling human-readable output.
func (f *HumanPrintFlags) ToPrinter(outputFormat string) (printers.ResourcePrinter, error) {
	if len(outputFormat) > 0 && outputFormat != "wide" {
		return nil, genericclioptions.NoCompatiblePrinterError{Options: f, AllowedFormats: f.AllowedFormats()}
	}

	decoder := codescheme.Codecs.UniversalDecoder()

	showKind := false
	if f.ShowKind != nil {
		showKind = *f.ShowKind
	}

	showLabels := false
	if f.ShowLabels != nil {
		showLabels = *f.ShowLabels
	}

	columnLabels := []string{}
	if f.ColumnLabels != nil {
		columnLabels = *f.ColumnLabels
	}

	p := NewHumanReadablePrinter(decoder, PrintOptions{
		Kind:          f.Kind,
		WithKind:      showKind,
		NoHeaders:     f.NoHeaders,
		Wide:          outputFormat == "wide",
		WithNamespace: f.WithNamespace,
		ColumnLabels:  columnLabels,
		ShowLabels:    showLabels,
	})
	printersinternal.AddHandlers(p)

	// TODO(juanvallejo): handle sorting here

	return p, nil
}

// NewHumanReadablePrinter creates a HumanReadablePrinter.
// If encoder and decoder are provided, an attempt to convert unstructured types to internal types is made.
func NewHumanReadablePrinter(decoder runtime.Decoder, options PrintOptions) *HumanReadablePrinter {
	printer := &HumanReadablePrinter{
		handlerMap: make(map[reflect.Type]*handlerEntry),
		options:    options,
		decoder:    decoder,
	}
	return printer
}

// ToPrinter attempts to find a composed set of PrintFlags suitable for
// returning a printer based on current flag values.
func (f *PrintFlags) ToPrinter() (printers.ResourcePrinter, error) {
	outputFormat := ""
	if f.OutputFormat != nil {
		outputFormat = *f.OutputFormat
	}

	noHeaders := false
	if f.NoHeaders != nil {
		noHeaders = *f.NoHeaders
	}
	f.HumanReadableFlags.NoHeaders = noHeaders
	f.CustomColumnsFlags.NoHeaders = noHeaders

	// for "get.go" we want to support a --template argument given, even when no --output format is provided
	if f.TemplateFlags.TemplateArgument != nil && len(*f.TemplateFlags.TemplateArgument) > 0 && len(outputFormat) == 0 {
		outputFormat = "go-template"
	}

	if p, err := f.TemplateFlags.ToPrinter(outputFormat); !genericclioptions.IsNoCompatiblePrinterError(err) {
		return p, err
	}

	if f.TemplateFlags.TemplateArgument != nil {
		f.CustomColumnsFlags.TemplateArgument = *f.TemplateFlags.TemplateArgument
	}

	if p, err := f.JSONYamlPrintFlags.ToPrinter(outputFormat); !genericclioptions.IsNoCompatiblePrinterError(err) {
		return p, err
	}

	if p, err := f.HumanReadableFlags.ToPrinter(outputFormat); !genericclioptions.IsNoCompatiblePrinterError(err) {
		return p, err
	}

	if p, err := f.NamePrintFlags.ToPrinter(outputFormat); !genericclioptions.IsNoCompatiblePrinterError(err) {
		return p, err
	}

	return nil, genericclioptions.NoCompatiblePrinterError{OutputFormat: &outputFormat, AllowedFormats: f.AllowedFormats()}
}


func PrintResult(infos []*resource.Info) {
	for _, info := range infos {
		fmt.Printf("Type : %s, id: %s\n", info.Object.GetObjectKind().GroupVersionKind().Kind, info.Name)
	}
}

func shouldGetNewPrinterForMapping(printer printers.ResourcePrinter, lastMapping, mapping *meta.RESTMapping) bool {
	return printer == nil || lastMapping == nil || mapping == nil || mapping.Resource != lastMapping.Resource
}

// OriginalPositioner and NopPositioner is required for swap/sort operations of data in table format
type OriginalPositioner interface {
	OriginalPosition(int) int
}

// GetNewTabWriter returns a tabwriter that translates tabbed columns in input into properly aligned text.
func GetNewTabWriter(output io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(output, tabwriterMinWidth, tabwriterWidth, tabwriterPadding, tabwriterPadChar, tabwriterFlags)
}


type Params struct {
	configFlags *genericclioptions.ConfigFlags
	PrintFlags  *PrintFlags
	builder     *resource.Builder
	Sort        bool
	Out         io.Writer
	ErrOut      io.Writer
	NoHeaders   bool
	PrintWithOpenAPICols   bool
	ToPrinter   func(*meta.RESTMapping, bool, bool) (printers.ResourcePrinterFunc, error)
}

type PrintFlags struct {
	JSONYamlPrintFlags *genericclioptions.JSONYamlPrintFlags
	NamePrintFlags     *genericclioptions.NamePrintFlags
	CustomColumnsFlags *CustomColumnsPrintFlags
	HumanReadableFlags *HumanPrintFlags
	TemplateFlags      *genericclioptions.KubeTemplatePrintFlags

	NoHeaders    *bool
	OutputFormat *string
}

// CustomColumnsPrintFlags provides default flags necessary for printing
// custom resource columns from an inline-template or file.
type CustomColumnsPrintFlags struct {
	NoHeaders        bool
	TemplateArgument string
}

// HumanPrintFlags provides default flags necessary for printing.
// Given the following flag values, a printer can be requested that knows
// how to handle printing based on these values.
type HumanPrintFlags struct {
	ShowKind     *bool
	ShowLabels   *bool
	SortBy       *string
	ColumnLabels *[]string

	// get.go-specific values
	NoHeaders bool

	Kind               schema.GroupKind
	AbsoluteTimestamps bool
	WithNamespace      bool
}

type PrintOptions struct {
	// supported Format types can be found in pkg/printers/printers.go
	OutputFormatType     string
	OutputFormatArgument string

	NoHeaders          bool
	WithNamespace      bool
	WithKind           bool
	Wide               bool
	ShowAll            bool
	ShowLabels         bool
	AbsoluteTimestamps bool
	Kind               schema.GroupKind
	ColumnLabels       []string

	SortBy string

	// indicates if it is OK to ignore missing keys for rendering an output template.
	AllowMissingKeys bool
}

type handlerEntry struct {
	columnDefinitions []metav1beta1.TableColumnDefinition
	printRows         bool
	printFunc         reflect.Value
	args              []reflect.Value
}

// HumanReadablePrinter is an implementation of ResourcePrinter which attempts to provide
// more elegant output. It is not threadsafe, but you may call PrintObj repeatedly; headers
// will only be printed if the object type changes. This makes it useful for printing items
// received from watches.
type HumanReadablePrinter struct {
	handlerMap     map[reflect.Type]*handlerEntry
	defaultHandler *handlerEntry
	options        PrintOptions
	lastType       interface{}
	skipTabWriter  bool
	encoder        runtime.Encoder
	decoder        runtime.Decoder
}

func NewParams(flags genericclioptions.ConfigFlags) *Params {
	return &Params{
		configFlags: genericclioptions.NewConfigFlags(),
	}
}

func DecodeIntoTable(obj runtime.Object) (runtime.Object, error) {
	if obj.GetObjectKind().GroupVersionKind().Kind != "Table" {
		return nil, fmt.Errorf("attempt to decode non-Table object into a v1beta1.Table")
	}

	unstr, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("attempt to decode non-Unstructured object")
	}
	table := &metav1beta1.Table{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.Object, table); err != nil {
		return nil, err
	}

	for i := range table.Rows {
		row := &table.Rows[i]
		if row.Object.Raw == nil || row.Object.Object != nil {
			continue
		}
		converted, err := runtime.Decode(unstructured.UnstructuredJSONScheme, row.Object.Raw)
		if err != nil {
			return nil, err
		}
		row.Object.Object = converted
	}

	return table, nil
}

func multipleGVKsRequested(infos []*resource.Info) bool {
	if len(infos) < 2 {
		return false
	}
	gvk := infos[0].Mapping.GroupVersionKind
	for _, info := range infos {
		if info.Mapping.GroupVersionKind != gvk {
			return true
		}
	}
	return false
}