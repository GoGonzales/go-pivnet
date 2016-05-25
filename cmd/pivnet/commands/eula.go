package commands

import (
	"fmt"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

type EULAsCommand struct {
}

type EULACommand struct {
	EULASlug string `long:"eula-slug" description:"EULA slug e.g. pivotal_software_eula" required:"true"`
}

type AcceptEULACommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

func (command *EULAsCommand) Execute([]string) error {
	client := NewClient()
	eulas, err := client.EULA.List()
	if err != nil {
		return err
	}

	return printEULAs(eulas)
}

func printEULA(eula pivnet.EULA) error {
	switch Pivnet.Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		eulaAsString := []string{
			strconv.Itoa(eula.ID), eula.Slug, eula.Name,
		}
		table.Append(eulaAsString)
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return Printer.PrintJSON(eula)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(eula)
	}

	return nil
}

func (command *EULACommand) Execute([]string) error {
	client := NewClient()
	eula, err := client.EULA.Get(command.EULASlug)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	return printEULA(eula)
}

func printEULAs(eulas []pivnet.EULA) error {
	switch Pivnet.Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		for _, e := range eulas {
			eulaAsString := []string{
				strconv.Itoa(e.ID), e.Slug, e.Name,
			}
			table.Append(eulaAsString)
		}
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return Printer.PrintJSON(eulas)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(eulas)
	}

	return nil
}

func (command *AcceptEULACommand) Execute([]string) error {
	client := NewClient()

	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return err
	}

	var release pivnet.Release
	for _, r := range releases {
		if r.Version == command.ReleaseVersion {
			release = r
			break
		}
	}

	if release.Version != command.ReleaseVersion {
		return fmt.Errorf("release not found")
	}

	err = client.EULA.Accept(command.ProductSlug, release.ID)
	if err != nil {
		return err
	}

	if Pivnet.Format == printer.PrintAsTable {
		_, err = fmt.Fprintf(
			OutputWriter,
			"eula acccepted successfully for %s/%s\n",
			command.ProductSlug,
			command.ReleaseVersion,
		)
	}

	return err
}
