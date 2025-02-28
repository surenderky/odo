package registry

import (
	"context"
	"errors"

	// Built-in packages
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	// Third-party packages

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/preference"
)

const listCommandName = "list"

// "odo preference registry list" command description and examples
var (
	listDesc = ktemplates.LongDesc(`List devfile registry`)

	listExample = ktemplates.Examples(`# List devfile registry
	%[1]s
	`)
)

// ListOptions encapsulates the options for "odo preference registry list" command
type ListOptions struct {
	// Clients
	clientset *clientset.Clientset
}

// NewListOptions creates a new ListOptions instance
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

func (o *ListOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes ListOptions after they've been created
func (o *ListOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	return nil
}

// Validate validates the ListOptions based on completed values
func (o *ListOptions) Validate() (err error) {
	return nil
}

// Run contains the logic for "odo preference registry list" command
func (o *ListOptions) Run(ctx context.Context) (err error) {
	registryList := o.clientset.PreferenceClient.RegistryList()
	if registryList == nil || len(*registryList) == 0 {
		//revive:disable:error-strings This is a top-level error message displayed as is to the end user
		return errors.New("No devfile registries added to the configuration. Refer `odo preference registry add -h` to add one")
		//revive:enable:error-strings
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "NAME", "\t", "URL", "\t", "SECURE")
	o.printRegistryList(w, registryList)
	w.Flush()
	return nil
}

func (o *ListOptions) printRegistryList(w io.Writer, registryList *[]preference.Registry) {
	if registryList == nil {
		return
	}

	regList := *registryList
	// Loop backwards here to ensure the registry display order is correct (display latest newly added registry firstly)
	for i := len(regList) - 1; i >= 0; i-- {
		registry := regList[i]
		secure := "No"
		if registry.Secure {
			secure = "Yes"
		}
		fmt.Fprintln(w, registry.Name, "\t", registry.URL, "\t", secure)
	}
}

// NewCmdList implements the "odo preference registry list" command
func NewCmdList(name, fullName string) *cobra.Command {
	o := NewListOptions()
	registryListCmd := &cobra.Command{
		Use:     name,
		Short:   listDesc,
		Long:    listDesc,
		Example: fmt.Sprintf(fmt.Sprint(listExample), fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(registryListCmd, clientset.PREFERENCE)
	return registryListCmd
}
