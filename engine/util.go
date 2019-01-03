// Mgmt
// Copyright (C) 2013-2018+ James Shubin and the project contributors
// Written by James Shubin <james@shubin.ca> and the project contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package engine

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// resourceRegisterCheck checks that each resource is safe to register. It is
// only called in tests and in the resource Register function. It checks that
// the exported fields don't contain any structs that contain unexported fields,
// because this can cause parts of our code to panic because reflect.StructOf is
// lame.
func resourceRegisterCheck(res Res) error {
	ts := reflect.TypeOf(res).Elem() // pointer to struct, then struct
	if k := ts.Kind(); k != reflect.Struct {
		return fmt.Errorf("expected struct, got: %s", k)
	}

	for i := 0; i < ts.NumField(); i++ {
		field := ts.Field(i)

		// skip embedded traits.* structs
		if strings.HasPrefix(field.Type.String(), "traits.") {
			continue
		}
		// skip top-level private fields
		if s := strings.ToUpper(field.Name[0:1]); s != field.Name[0:1] {
			continue
		}
		// check the nested type!
		if err := structCheck(field.Type); err != nil {
			return err
		}
	}

	return nil
}

// structCheck is a helper function for recursive checking of structs. It errors
// if it finds a struct field that is unexported.
func structCheck(st reflect.Type) error {
	switch st.Kind() {
	case reflect.Array, reflect.Slice:
		return structCheck(st.Elem())

	case reflect.Map:
		if err := structCheck(st.Key()); err != nil {
			return err
		}
		return structCheck(st.Elem())

	case reflect.Ptr:
		return structCheck(st.Elem()) // TODO: is this correct?

	case reflect.Struct:
		// handle struct at the end
	default:
		return nil // we don't need to handle these
	}

	// check the struct
	for j := 0; j < st.NumField(); j++ { // new counter var for fun
		name := st.Field(j).Name
		if s := strings.ToUpper(name[0:1]); s != name[0:1] {
			return fmt.Errorf("private field found: %s", name)
		}

		// recurse
		if err := structCheck(st.Field(j).Type); err != nil {
			return err
		}
	}

	return nil
}

// ResourceSlice is a linear list of resources. It can be sorted.
type ResourceSlice []Res

func (rs ResourceSlice) Len() int           { return len(rs) }
func (rs ResourceSlice) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }
func (rs ResourceSlice) Less(i, j int) bool { return rs[i].String() < rs[j].String() }

// Sort the list of resources and return a copy without modifying the input.
func Sort(rs []Res) []Res {
	resources := []Res{}
	for _, r := range rs { // copy
		resources = append(resources, r)
	}
	sort.Sort(ResourceSlice(resources))
	return resources
	// sort.Sort(ResourceSlice(rs)) // this is wrong, it would modify input!
	//return rs
}
