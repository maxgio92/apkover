package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	melange "chainguard.dev/melange/pkg/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/maxgio92/apkover/internal/output"
	"github.com/maxgio92/apkover/internal/utils"
	gopipeline "github.com/maxgio92/apkover/pkg/pipeline/go"
	rustpipeline "github.com/maxgio92/apkover/pkg/pipeline/rust"
	"github.com/maxgio92/apkover/pkg/report"
)

var (
	errCovNotFoundFromLog = errors.New("cannot extract coverage from log")
	errCovLow             = errors.New("test coverage is below the minimum required")
	coverage              float64
)

type Options struct {
	ConfigPath string
	Language   string
	LogLevel   string
	MinCov     float64
	Output     string
}

func NewCmd() *cobra.Command {
	o := new(Options)

	cmd := &cobra.Command{
		Use:           usage,
		Short:         description,
		RunE:          o.Run,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.Flags().StringVarP(&o.ConfigPath, "config", "c", "", "path to package config file")
	cmd.Flags().StringVar(&o.Language, "language", "go", "main language of the package")
	cmd.Flags().StringVar(&o.LogLevel, "log-level", "info", "log level")
	cmd.Flags().Float64Var(&o.MinCov, "fail-under", 0, "The minimum accepted coverage, expressed as percentage (e.g. 80 for 80% of coverage). Fail if it's under the specified threshold.")
	cmd.Flags().StringVarP(&o.Output, "output", "o", "text", "output format (text, json, yaml)")
	cmd.MarkFlagRequired("config")

	return cmd
}

func (o *Options) Run(_ *cobra.Command, _ []string) error {
	// Configure zerolog to use ConsoleWriter (for color output).
	// TODO: remove the logger responsibility from the command.
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	logLevel, err := zerolog.ParseLevel(o.LogLevel)
	if err != nil {
		return errors.Wrap(err, "could not parse log level")
	}
	log.Logger = zerolog.New(consoleWriter).Level(logLevel).With().Timestamp().Logger()

	file, err := os.Open(o.ConfigPath)
	if err != nil {
		return errors.Wrap(err, "error opening file")
	}
	defer file.Close()

	cfg, err := melange.ParseConfiguration(context.Background(), o.ConfigPath)
	if err != nil {
		return err
	}

	pkgName := cfg.Name()

	// Update the build pipeline.
	log.Info().Str("package", pkgName).Msg("Updating the build pipeline to instrument the package")

	err = updateBuildPipeline(cfg, o.Language)
	if err != nil {
		return err
	}

	// Update the test pipeline.
	log.Info().Str("package", pkgName).Msg("Updating the test pipeline to generate coverage data")

	err = updateTest(cfg.Test, o.Language)
	if err != nil {
		return err
	}

	// Write the config back to disk.
	log.Info().Str("package", pkgName).Msg("Writing the package config to disk")

	cfgB, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	err = os.WriteFile(o.ConfigPath, cfgB, 0644)
	if err != nil {
		return err
	}

	// Build the package.
	log.Info().Str("package", pkgName).Int("steps", len(cfg.Pipeline)).Msg("Re-building the package instrumented")
	// TODO: abstract Melange runner.
	err = runMelange(
		filepath.Dir(o.ConfigPath),
		fmt.Sprintf("%s/%s", "package", pkgName),
		fmt.Sprintf("MELANGE_EXTRA_OPTS='--runner=%s'", melangeRunner),
	)
	if err != nil {
		return err
	}

	// Test the package.
	log.Info().Str("package", pkgName).Int("steps", len(cfg.Test.Pipeline)).Msg("Running tests and writing coverage data")
	err = runMelange(
		filepath.Dir(o.ConfigPath),
		fmt.Sprintf("%s/%s", "test", pkgName),
		fmt.Sprintf("MELANGE_EXTRA_OPTS='--runner=%s'", melangeRunner),
	)
	if err != nil {
		return err
	}

	// Print the report into the desired format.
	rprt := report.NewCovReport(
		report.WithPackageName(pkgName),
		report.WithPackageVersion(cfg.Package.Version),
		report.WithPackageEpoch(cfg.Package.Epoch),
		report.WithCovPercent(coverage),
	)
	switch o.Output {
	case "text":
		output.PrettyPercentageTest(int(math.Ceil(rprt.CovFloat*100)), int(o.MinCov), "Test Coverage")
	case "json":
		b, err := json.Marshal(rprt)
		if err != nil {
			return errors.Wrap(err, "error marshaling report to JSON")
		}
		fmt.Println(string(b))
	case "yaml":
		b, err := yaml.Marshal(rprt)
		if err != nil {
			return errors.Wrap(err, "error marshaling report to YAML")
		}
		fmt.Println(string(b))
	default:
		output.PrettyPercentageTest(int(math.Ceil(rprt.CovFloat*100)), int(o.MinCov), "Test Coverage")
	}

	// Coverage is too low.
	if int(coverage*100) < int(o.MinCov) {
		return errCovLow
	}

	// Coverage is enough.
	return nil
}

// updateBuildPipeline updates the language specific build pipelines to generate
// coverage measurement instrumented artifacts like executable binaries.
func updateBuildPipeline(build *melange.Configuration, lang string) error {
	switch lang {
	case "go":
		return gopipeline.UpdateBuild(build)
	case "rust":
		return rustpipeline.UpdateBuild(build)
	default:
		return gopipeline.UpdateBuild(build)
	}
}

func updateTest(test *melange.Test, lang string) error {
	switch lang {
	case "go":
		return gopipeline.UpdateTest(test)
	case "rust":
		return rustpipeline.UpdateTest(test)
	default:
		return gopipeline.UpdateTest(test)
	}
}

// runMelange runs Melange via the Wolfi Make targets.
func runMelange(dir, target string, env ...string) error {
	var args []string
	for _, v := range env {
		args = append(args, "-e", v)
	}

	args = append(args, "-C", dir)
	args = append(args, target)

	stdout, stderr, errCh := utils.RunCmd("make", args...)

	// Melange output parsing error.
	var err error
	// Melange parsers wait group.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case line, ok := <-stdout:
				if !ok {
					stdout = nil
				} else {
					parseMelangeStdout(line)
				}
			case line, ok := <-stderr:
				if !ok {
					stderr = nil
				} else {
					err = parseMelangeStderr(line)
				}
			}
			// Stdout/stderr pipes from Melange command have been closed.
			if stdout == nil && stderr == nil {
				break
			}
		}
	}()

	// Wait for the Melange command to complete.
	if err := <-errCh; err != nil {
		return errors.Wrapf(err, "error running make %s", target)
	}
	// Check if there have been any errors parsing the Melange output.
	if err != nil {
		return errors.Wrap(err, "error parsing Melange output")
	}

	// Wait to finish parsing the Melange output.
	wg.Wait()

	return nil
}

// Stdout from Melange is just sent to the debug logs.
func parseMelangeStdout(text string) {
	log.Debug().Msg(text)
}

// Stderr from Melange contains the most meaningful information.
// In particular it contains stdout from the pipeline steps, and
// we want to capture the output from the coverage data analysis.
// TODO: abstract log parser from coverage collector.
func parseMelangeStderr(text string) error {
	// Retrieve coverage percentage from stdout.
	if strings.Contains(text, fmt.Sprintf("INFO %s", report.ReportCovPrefix)) {
		cov, err := extractCovFromLog(text)
		if err != nil {
			log.Error().Err(err).Msg("error extracting coverage")
			return err
		}
		coverage = cov / 100

		return nil
	}

	// Infer error logs.
	if strings.Contains(text, "ERRO") ||
		strings.Contains(strings.ToLower(text), "error") {
		log.Error().Msg(text)
	} else {
		log.Debug().Msg(text)
	}

	return nil
}

func extractCovFromLog(text string) (float64, error) {
	re := regexp.MustCompile(fmt.Sprintf(`%s (\d+\.\d+)%%`, report.ReportCovPrefix))
	match := re.FindStringSubmatch(text)
	if len(match) <= 1 {
		log.Warn().Msgf("cannot extract coverage percentage from log: %s", text)
		return coverage, errCovNotFoundFromLog
	}

	log.Debug().Msgf("Extracting coverage data from logs")
	cov, err := strconv.ParseFloat(match[1], 32)
	if err != nil {
		return coverage, err
	}

	return cov, nil
}
