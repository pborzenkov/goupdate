package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"debug/gosym"
)

type objInfo struct {
	symtab   []byte
	pclntab  []byte
	textAddr uint64
}

var (
	goPathSrc string
	goPathBin string

	errNotAGoBinary = fmt.Errorf("not a Go built binary")

	verbose = flag.Bool("verbose", false, "Print debug messages")
	force   = flag.Bool("force", false, "Don't ask for confirmation")
)

func init() {
	goPath, ok := os.LookupEnv("GOPATH")
	if !ok {
		fmt.Fprintf(os.Stderr, "error: GOPATH is not set\n")
		os.Exit(1)
	}
	goPathSrc = filepath.Join(goPath, "src")
	goPathBin = filepath.Join(goPath, "bin")
}

func debug(format string, a ...interface{}) {
	if !*verbose {
		return
	}
	fmt.Printf(format, a...)
}

func processBinary(file string, ask bool) error {
	if !filepath.IsAbs(file) {
		file = filepath.Join(goPathBin, file)
	}

	debug("Processing '%s'...", file)
	obj, err := getObjInfo(file)
	if err != nil {
		debug("FAIL, %v\n", err)
		return err
	}

	pcln := gosym.NewLineTable(obj.pclntab, obj.textAddr)
	tab, err := gosym.NewTable(obj.symtab, pcln)
	if err != nil {
		debug("FAIL, %v\n", err)
		return err
	}

	fun := tab.LookupFunc("main.main")
	srcFile, _, _ := tab.PCToLine(fun.Entry)

	if !strings.HasPrefix(srcFile, goPathSrc) {
		err := fmt.Errorf("binary '%s' wasn't built from current GOPATH", file)
		debug("FAIL, %v\n", err)
		return err
	}

	if ask {
		debug("\n")

		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Printf("Do you want to update '%s'? (y/n): ", file)
			answer, err := reader.ReadString('\n')
			if err != nil {
				debug("FAIL, %v\n", err)
				return err
			}
			answer = strings.ToLower(strings.TrimSpace(answer))
			if answer == "y" || answer == "yes" {
				break
			}
			if answer == "n" || answer == "no" {
				debug("SKIP\n")
				return nil
			}
		}
	}

	err = updateBinary(filepath.Dir(srcFile[len(goPathSrc)+1:]))
	if err != nil {
		debug("FAIL, %v\n", err)
		return err
	}
	debug("OK\n")
	return nil
}

func updateBinary(binary string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("go", "get", "-u", binary)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go get -u %s failed\n%v\n%s\n", binary, err, stderr.String())
	}

	cmd = exec.Command("go", "install", binary)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go install %s failed\n%v\n%s\n", binary, err, stderr.String())
	}
	return nil
}

func processAllBinaries() {
	filepath.Walk(goPathBin, func(path string, info os.FileInfo, err error) error {
		if info.Mode()&os.ModeType != 0 {
			return nil
		}

		err = processBinary(path, !*force)
		if err != nil {
			if err != errNotAGoBinary {
				fmt.Printf("Failed to process '%s': %v\n", path, err)
			}
		} else {
			fmt.Printf("Updated '%s'\n", path)
		}
		return nil
	})
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		processAllBinaries()
	}

	for _, f := range flag.Args() {
		err := processBinary(f, true)
		if err != nil && !*verbose {
			fmt.Printf("Failed to process '%s': %v\n", f, err)
		}
	}
}
