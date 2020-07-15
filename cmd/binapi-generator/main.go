// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen"
	"git.fd.io/govpp.git/binapigen/vppapi"
	"git.fd.io/govpp.git/internal/version"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION] API_FILES\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Parse API_FILES and generate Go bindings based on the options given:")
		flag.PrintDefaults()
	}
}

func main() {
	var (
		theApiDir        = flag.String("input-dir", vppapi.DefaultDir, "Input directory containing API files.")
		theInputFile     = flag.String("input-file", "", "DEPRECATED: Use program arguments to define files to generate.")
		theOutputDir     = flag.String("output-dir", "binapi", "Output directory where code will be generated.")
		importPrefix     = flag.String("import-prefix", "", "Define import path prefix to be used to import types.")
		generatorPlugins = flag.String("gen", "rpc", "List of generator plugins to run.")

		printVersion = flag.Bool("version", false, "Prints version and exits.")
		verboseLog   = flag.Bool("verbose", false, "Enable verbose logging.")
	)
	flag.Parse()

	if *printVersion {
		fmt.Fprintln(os.Stdout, version.Info())
		os.Exit(0)
	}

	if debug := os.Getenv("DEBUG_GOVPP"); strings.Contains(debug, "binapigen") {
		logrus.SetLevel(logrus.DebugLevel)
	} else if *verboseLog {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	var filesToGenerate []string
	if *theInputFile != "" {
		if flag.NArg() > 0 {
			fmt.Fprintln(os.Stderr, "input-file cannot be combined with files to generate in arguments")
			os.Exit(1)
		}
		filesToGenerate = append(filesToGenerate, *theInputFile)
	} else {
		filesToGenerate = append(filesToGenerate, flag.Args()...)
	}

	opts := binapigen.Options{
		ImportPrefix: *importPrefix,
		OutputDir:    *theOutputDir,
	}
	if opts.OutputDir == "binapi" {
		if wd, _ := os.Getwd(); filepath.Base(wd) == "binapi" {
			opts.OutputDir = "."
		}
	}
	apiDir := *theApiDir
	genPlugins := strings.Split(*generatorPlugins, ",")

	binapigen.Run(apiDir, filesToGenerate, opts, func(gen *binapigen.Generator) error {
		for _, file := range gen.Files {
			if !file.Generate {
				continue
			}
			binapigen.GenerateAPI(gen, file)
			for _, p := range genPlugins {
				if err := binapigen.RunPlugin(p, gen, file); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
