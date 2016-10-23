// +build linux

package main

import (
	"debug/elf"
)

func getObjInfo(file string) (objInfo, error) {
	var obj objInfo

	f, err := elf.Open(file)
	if err != nil {
		if _, ok := err.(*elf.FormatError); ok {
			return obj, errNotAGoBinary
		}
		return obj, err
	}
	defer f.Close()

	symtab := f.Section(".gosymtab")
	if symtab == nil {
		return obj, errNotAGoBinary
	}
	obj.symtab, err = symtab.Data()
	if err != nil {
		return obj, err
	}

	obj.pclntab, err = f.Section(".gopclntab").Data()
	if err != nil {
		return obj, err
	}

	obj.textAddr = f.Section(".text").Addr

	return obj, nil
}
