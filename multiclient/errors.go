// Copyright 2018 AMIS Technologies
// This file is part of the hypereth library.
//
// The hypereth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The hypereth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the hypereth library. If not, see <http://www.gnu.org/licenses/>.

package multiclient

import "strings"

type ClientError struct {
	client string
	err    error
}

func NewClientError(client string, err error) *ClientError {
	return &ClientError{
		client: client,
		err:    err,
	}
}

func (e *ClientError) Error() string {
	return e.err.Error()
}

func (e *ClientError) Client() string {
	return e.client
}

type MultipleError struct {
	errs []error
}

func NewMultipleError(errs []error) *MultipleError {
	return &MultipleError{
		errs: errs,
	}
}

func (e *MultipleError) Error() string {
	errstrs := make([]string, len(e.errs))
	for i, err := range e.errs {
		errstrs[i] = err.Error()
	}

	return strings.Join(errstrs, ",")
}

func (e *MultipleError) GetErrors() []error {
	return e.errs
}
