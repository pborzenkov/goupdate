// +build darwin

package main

import (
	"debug/macho"
)

func getObjInfo(file string) (objInfo, error) {
	var obj objInfo

	f, err := macho.Open(file)
	if err != nil {
		if _, ok := err.(*macho.FormatError); ok {
			return obj, errNotAGoBinary
		}
		return obj, err
	}
	defer f.Close()

	symtab := f.Section("__gosymtab")
	if symtab == nil {
		return obj, errNotAGoBinary
	}
	obj.symtab, err = symtab.Data()
	if err != nil {
		return obj, err
	}

	obj.pclntab, err = f.Section("__gopclntab").Data()
	if err != nil {
		return obj, err
	}

	obj.textAddr = f.Section("__text").Addr

	return obj, nil
}
