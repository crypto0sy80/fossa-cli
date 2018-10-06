package analyze

import (
	"fmt"

	"github.com/fatih/color"
	wordwrap "github.com/mitchellh/go-wordwrap"
	"github.com/urfave/cli"

	"github.com/apex/log"
	"github.com/fossas/fossa-cli/analyzers"
	"github.com/fossas/fossa-cli/api/fossa"
	"github.com/fossas/fossa-cli/cmd/fossa/display"
	"github.com/fossas/fossa-cli/cmd/fossa/flags"
	"github.com/fossas/fossa-cli/cmd/fossa/setup"
	"github.com/fossas/fossa-cli/config"
	"github.com/fossas/fossa-cli/errors"
	"github.com/fossas/fossa-cli/module"
	"github.com/fossas/fossa-cli/pkg"
)

var ShowOutput = "output"

var Cmd = cli.Command{
	Name:      "analyze",
	Usage:     "Analyze built dependencies",
	Action:    Run,
	ArgsUsage: "MODULE",
	Flags: flags.WithGlobalFlags(flags.WithAPIFlags(flags.WithOptions([]cli.Flag{
		cli.BoolFlag{Name: "show-output, output, o", Usage: "print results to stdout instead of uploading to FOSSA"},
		flags.TemplateF,
	}))),
}

var _ cli.ActionFunc = Run

func Run(ctx *cli.Context) error {
	err := setup.SetContext(ctx)
	if err != nil {
		log.Fatalf("Could not initialize: %s", err.Error())
	}

	if !ctx.Bool(ShowOutput) {
		err = fossa.SetAPIKey(config.APIKey())
		switch err {
		case fossa.ErrMissingAPIKey:
			return &errors.Error{
				Code:    "E_MISSING_API_KEY",
				Type:    errors.UserInput,
				Message: "A FOSSA API key is needed to run this command.",
				Troubleshooting: `
` + wordwrap.WrapString("Running `fossa analyze` performs a dependency analysis and uploads the result to FOSSA. To run an analysis without uploading results, run:", 78) + `

    ` + color.HiGreenString("fossa analyze --output") + `

` + wordwrap.WrapString("You can provide your API key by setting the $FOSSA_API_KEY environment variable. For example, try running:", 78) + `

    ` + color.HiGreenString("FOSSA_API_KEY=<YOUR_API_KEY_HERE> $command") + `

` + wordwrap.WrapString("You can create an API key for your FOSSA account at:", 78) + `

    ` + color.HiBlueString("$endpoint/account/settings/integrations/api_tokens") + `
`,
			}
		default:
			return err
		}
	}

	modules, err := config.Modules()
	if err != nil {
		log.Fatalf("Could not parse modules: %s", err.Error())
	}
	if len(modules) == 0 {
		log.Fatal("No modules specified.")
	}

	analyzed, err := Do(modules)
	if err != nil {
		log.Fatalf("Could not analyze modules: %s", err.Error())
		return err
	}

	log.Debugf("analyzed: %#v", analyzed)
	normalized, err := fossa.Normalize(analyzed)
	if err != nil {
		log.Fatalf("Could not normalize output: %s", err.Error())
		return err
	}

	if ctx.Bool(ShowOutput) {
		if tmplFile := ctx.String(flags.Template); tmplFile != "" {
			output, err := display.TemplateFile(tmplFile, normalized)
			fmt.Println(output)
			if err != nil {
				log.Fatalf("Could not parse template data: %s", err.Error())
			}
		} else {
			_, err := display.JSON(normalized)
			if err != nil {
				log.Fatalf("Could not serialize to JSON: %s", err.Error())
			}
		}

		return nil
	}

	return uploadAnalysis(normalized)
}

func Do(modules []module.Module) (analyzed []module.Module, err error) {
	defer display.ClearProgress()
	for i, m := range modules {
		display.InProgress(fmt.Sprintf("Analyzing module (%d/%d): %s", i+1, len(modules), m.Name))

		// Handle raw modules differently from all others.
		// TODO: maybe this should occur during the analysis step?
		// TODO: maybe this should target a third-party folder, rather than a single
		// folder? Maybe "third-party folder" should be a separate module type?
		if m.Type == pkg.Raw {
			locator, err := fossa.UploadTarball(m.BuildTarget)
			if err != nil {
				log.Warnf("Could not upload raw module: %s", err.Error())
			}
			id := pkg.ID{
				Type:     pkg.Raw,
				Name:     locator.Project,
				Revision: locator.Revision,
			}
			m.Imports = []pkg.Import{pkg.Import{Resolved: id}}
			m.Deps = make(map[pkg.ID]pkg.Package)
			m.Deps[id] = pkg.Package{
				ID: id,
			}
			analyzed = append(analyzed, m)
			continue
		}

		analyzer, err := analyzers.New(m)
		if err != nil {
			analyzed = append(analyzed, m)
			log.Warnf("Could not load analyzer: %s", err.Error())
			continue
		}
		built, err := analyzer.IsBuilt()
		if err != nil {
			log.Warnf("Could not determine whether module is built: %s", err.Error())
		}
		if !built {
			log.Warnf("Module does not appear to be built")
		}
		deps, err := analyzer.Analyze()
		if err != nil {
			log.Fatalf("Could not analyze: %s", err.Error())
		}
		m.Imports = deps.Direct
		m.Deps = deps.Transitive
		analyzed = append(analyzed, m)
	}
	display.ClearProgress()

	return analyzed, err
}

func uploadAnalysis(normalized []fossa.SourceUnit) error {
	display.InProgress("Uploading analysis...")
	locator, err := fossa.Upload(
		config.Title(),
		fossa.Locator{
			Fetcher:  config.Fetcher(),
			Project:  config.Project(),
			Revision: config.Revision(),
		},
		fossa.UploadOptions{
			Branch:         config.Branch(),
			ProjectURL:     config.ProjectURL(),
			JIRAProjectKey: config.JIRAProjectKey(),
			Link:           config.Link(),
			Team:           config.Team(),
		},
		normalized)
	display.ClearProgress()
	if err != nil {
		log.Fatalf("Error during upload: %s", err.Error())
		return err
	}
	fmt.Println(locator.ReportURL())
	return nil
}
