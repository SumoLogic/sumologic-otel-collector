// Copyright 2021, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package activedirectoryinvreceiver

import (
	"runtime"

	adsi "github.com/go-adsi/adsi"
)

// Container is an interface for an Active Directory container
type Container interface {
	ToObject() (Object, error)
	Close()
	Children() (ObjectIter, error)
}

// ADSIContainer is a wrapper for an Active Directory container
type ADSIContainer struct {
	windowsADContainer *adsi.Container
}

// ToObject converts an Active Directory container to an Active Directory object
func (c *ADSIContainer) ToObject() (Object, error) {
	object, err := c.windowsADContainer.ToObject()
	if err != nil {
		return nil, err
	}
	return &ADObject{object}, nil
}

// Close closes an Active Directory container
func (c *ADSIContainer) Close() {
	c.windowsADContainer.Close()
}

// Children returns the children of an Active Directory container
func (c *ADSIContainer) Children() (ObjectIter, error) {
	objectIter, err := c.windowsADContainer.Children()
	if err != nil {
		return nil, err
	}
	return &ADObjectIter{objectIter}, nil
}

// Object is an interface for an Active Directory object
type Object interface {
	Attrs(key string) ([]interface{}, error)
	ToContainer() (Container, error)
}

// ADObject is a wrapper for an Active Directory object
type ADObject struct {
	windowsADObject *adsi.Object
}

// Attrs returns the attributes of an Active Directory object
func (o *ADObject) Attrs(key string) ([]interface{}, error) {
	return o.windowsADObject.Attr(key)
}

// ToContainer converts an Active Directory object to an Active Directory container
func (o *ADObject) ToContainer() (Container, error) {
	container, err := o.windowsADObject.ToContainer()
	if err != nil {
		return nil, err
	}
	return &ADSIContainer{container}, nil
}

// ObjectIter is an interface for an Active Directory object iterator
type ObjectIter interface {
	Next() (*adsi.Object, error)
	Close()
}

// ADObjectIter is a wrapper for an Active Directory object iterator
type ADObjectIter struct {
	windowsADObjectIter *adsi.ObjectIter
}

// Next returns the next Active Directory object in the iterator
func (o *ADObjectIter) Next() (*adsi.Object, error) {
	return o.windowsADObjectIter.Next()
}

// Close closes an Active Directory object iterator
func (o *ADObjectIter) Close() {
	o.windowsADObjectIter.Close()
}

// RuntimeInfo is an interface for runtime information
type RuntimeInfo interface {
	SupportedOS() bool
}

// ADRuntimeInfo is a wrapper for runtime information
type ADRuntimeInfo struct{}

// SupportedOS returns whether the runtime is supported
func (r *ADRuntimeInfo) SupportedOS() bool {
	return (runtime.GOOS == "windows")
}
