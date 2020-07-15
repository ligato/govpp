//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package binapigen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/binapigen/vppapi"
)

func Run(apiDir string, filesToGenerate []string, opts Options, f func(*Generator) error) {
	if err := run(apiDir, filesToGenerate, opts, f); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
}

func run(apiDir string, filesToGenerate []string, opts Options, fn func(*Generator) error) error {
	apifiles, err := vppapi.ParseDir(apiDir)
	if err != nil {
		return err
	}

	gen, err := New(opts, apifiles, filesToGenerate)
	if err != nil {
		return err
	}
	gen.vppVersion = vppapi.ResolveVPPVersion(apiDir)

	if fn == nil {
		GenerateDefault(gen)
	} else {
		if err := fn(gen); err != nil {
			return err
		}
	}

	if err = gen.Generate(); err != nil {
		return err
	}

	return nil
}

func GenerateDefault(gen *Generator) {
	for _, file := range gen.Files {
		if !file.Generate {
			continue
		}
		GenerateAPI(gen, file)
		GenerateRPC(gen, file)
	}
}

var Logger = logrus.New() //*logrus.Logger

func init() {
	if debug := os.Getenv("DEBUG_GOVPP"); strings.Contains(debug, "binapigen") {
		Logger.SetLevel(logrus.DebugLevel)
		logrus.SetLevel(logrus.DebugLevel)
	} else if debug != "" {
		Logger.SetLevel(logrus.InfoLevel)
	} else {
		Logger.SetLevel(logrus.WarnLevel)
	}
}

func logf(f string, v ...interface{}) {
	Logger. /*WithField("logger", "binapigen").*/ Debugf(f, v...)
}
