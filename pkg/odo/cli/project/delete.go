package project

import (
	"context"
	"fmt"

	odoerrors "github.com/redhat-developer/odo/pkg/errors"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/spf13/cobra"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const deleteRecommendedCommandName = "delete"

var (
	deleteExample = ktemplates.Examples(`
	# Delete a project
	%[1]s myproject  
	`)

	deleteLongDesc = ktemplates.LongDesc(`Delete a project and all resources deployed in the project being deleted.
	This command directly performs actions on the cluster and doesn't require a push.
	`)

	deleteShortDesc = `Delete a project`
)

// ProjectDeleteOptions encapsulates the options for the odo project delete command
type ProjectDeleteOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Parameters
	projectName string

	// Flags
	forceFlag bool
	waitFlag  bool
}

// NewProjectDeleteOptions creates a ProjectDeleteOptions instance
func NewProjectDeleteOptions() *ProjectDeleteOptions {
	return &ProjectDeleteOptions{}
}

func (o *ProjectDeleteOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes ProjectDeleteOptions after they've been created
func (pdo *ProjectDeleteOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	pdo.projectName = args[0]
	pdo.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	return err
}

// Validate validates the parameters of the ProjectDeleteOptions
func (pdo *ProjectDeleteOptions) Validate() error {
	// Validate existence of the project to be deleted
	isValidProject, err := pdo.clientset.ProjectClient.Exists(pdo.projectName)
	if kerrors.IsForbidden(err) {
		return &odoerrors.Unauthorized{}
	}
	if !isValidProject {
		//revive:disable:error-strings This is a top-level error message displayed as is to the end user
		return fmt.Errorf("The project %q does not exist. Please check the list of projects using `odo project list`", pdo.projectName)
		//revive:enable:error-strings
	}
	return nil
}

// Run the project delete command
func (pdo *ProjectDeleteOptions) Run(ctx context.Context) (err error) {

	// Create the "spinner"
	s := &log.Status{}

	// This to set the project in the file and runtime
	err = pdo.clientset.ProjectClient.SetCurrent(pdo.projectName)
	if err != nil {
		return err
	}

	// Prints out what will be deleted
	// This function doesn't support devfile components.
	// TODO: fix this once we have proper abstraction layer on top of devfile components
	//err = printDeleteProjectInfo(pdo.Context, pdo.projectName)
	//if err != nil {
	//	return err
	//}

	if pdo.forceFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete project %v", pdo.projectName)) {

		// If the --wait parameter has been passed, we add a spinner..
		if pdo.waitFlag {
			s = log.Spinner("Waiting for project to be deleted")
			defer s.End(false)
		}

		err := pdo.clientset.ProjectClient.Delete(pdo.projectName, pdo.waitFlag)
		if err != nil {
			return err
		}
		s.End(true)

		successMessage := fmt.Sprintf("Deleted project : %v", pdo.projectName)
		log.Success(successMessage)
		log.Warning("Warning! Projects are asynchronously deleted from the cluster. odo does its best to delete the project. Due to multi-tenant clusters, the project may still exist on a different node.")

		return nil
	}

	return fmt.Errorf("aborting deletion of project: %v", pdo.projectName)
}

// NewCmdProjectDelete creates the project delete command
func NewCmdProjectDelete(name, fullName string) *cobra.Command {
	o := NewProjectDeleteOptions()

	projectDeleteCmd := &cobra.Command{
		Use:     name,
		Short:   deleteShortDesc,
		Long:    deleteLongDesc,
		Example: fmt.Sprintf(deleteExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(projectDeleteCmd, clientset.PROJECT)

	projectDeleteCmd.Flags().BoolVarP(&o.waitFlag, "wait", "w", false, "Wait until the project has been completely deleted")
	projectDeleteCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete project without prompting")

	return projectDeleteCmd
}
