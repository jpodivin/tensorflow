/*
Copyright 2016 The TensorFlow Authors. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package op

import (
	"strings"
	"testing"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

func TestAddGradients(t *testing.T) {
	var (
		s  = NewScope()
		x1 = Placeholder(s.SubScope("x1"), tf.Float)
		x2 = Placeholder(s.SubScope("x2"), tf.Float)
		y0 = Square(s.SubScope("y0"), x1)
		y1 = Square(s.SubScope("y1"), y0)
		y2 = AddN(s.SubScope("y2"), []tf.Output{y0, x2})
	)

	grads0 := Gradients(s, "", []tf.Output{y1}, []tf.Output{x1})
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if len(grads0) != 1 {
		t.Fatal(len(grads0))
	}
	if grads0[0].DataType() != tf.Float {
		t.Fatalf("Got DataType %v, wanted %v", grads0[0].DataType(), tf.Float)
	}

	grads1 := Gradients(s, "", []tf.Output{y2}, []tf.Output{x1, x2})
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if len(grads1) != 2 {
		t.Fatal(len(grads1))
	}
	if grads1[0].DataType() != tf.Float {
		t.Fatalf("Got DataType %v, wanted %v", grads1[0].DataType(), tf.Float)
	}
	if grads1[1].DataType() != tf.Float {
		t.Fatalf("Got DataType %v, wanted %v", grads1[1].DataType(), tf.Float)
	}

	graph, err := s.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	sess, err := tf.NewSession(graph, nil)
	if err != nil {
		t.Fatal(err)
	}

	c1, _ := tf.NewTensor(float32(3.0))
	c2, _ := tf.NewTensor(float32(3.0))
	outputs, err := sess.Run(
		map[tf.Output]*tf.Tensor{x1: c1, x2: c2},
		[]tf.Output{grads0[0], grads1[0], grads1[1]},
		nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(outputs) != 3 {
		t.Fatal(len(outputs))
	}
	if outputs[0].Value().(float32) != 108.0 {
		t.Fatalf("Got %v, wanted float 108.0", outputs[0].Value())
	}
	if outputs[1].Value().(float32) != 6.0 {
		t.Fatalf("Got %v, wanted float 6.0", outputs[1].Value())
	}
	if outputs[2].Value().(float32) != 1.0 {
		t.Fatalf("Got %v, wanted float 1.0", outputs[2].Value())
	}
}

func TestAddGradientsSums(t *testing.T) {
	var (
		s  = NewScope()
		x  = Placeholder(s.SubScope("x"), tf.Float)
		y0 = Square(s.SubScope("y0"), x)
		y1 = Square(s.SubScope("y1"), y0)
	)

	grad := Gradients(s, "", []tf.Output{y0, y1}, []tf.Output{x})
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if len(grad) != 1 {
		t.Fatal(len(grad))
	}
	if grad[0].DataType() != tf.Float {
		t.Fatalf("Got DataType %v, wanted %v", grad[0].DataType(), tf.Float)
	}

	graph, err := s.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	sess, err := tf.NewSession(graph, nil)
	if err != nil {
		t.Fatal(err)
	}

	c, _ := tf.NewTensor(float32(3.0))
	outputs, err := sess.Run(
		map[tf.Output]*tf.Tensor{x: c},
		[]tf.Output{grad[0]},
		nil)
	if err != nil {
		t.Fatal(err)
	}
	if outputs[0].Value().(float32) != 114.0 {
		t.Fatalf("Got %v, wanted float 114.0", outputs[0].Value())
	}
}

func TestAddGradientsWithInitialValues(t *testing.T) {
	var (
		s  = NewScope()
		x  = Placeholder(s.SubScope("x1"), tf.Float)
		y0 = Square(s.SubScope("y0"), x)
		y1 = Square(s.SubScope("y1"), y0)
	)

	grads0 := Gradients(s, "", []tf.Output{y1}, []tf.Output{y0})
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if len(grads0) != 1 {
		t.Fatal(len(grads0))
	}
	if grads0[0].DataType() != tf.Float {
		t.Fatalf("Got DataType %v, wanted %v", grads0[0].DataType(), tf.Float)
	}

	grads1 := Gradients(s, "", []tf.Output{y0}, []tf.Output{x}, grads0[0])
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if len(grads1) != 1 {
		t.Fatal(len(grads1))
	}
	if grads1[0].DataType() != tf.Float {
		t.Fatalf("Got DataType %v, wanted %v", grads1[0].DataType(), tf.Float)
	}

	graph, err := s.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	sess, err := tf.NewSession(graph, nil)
	if err != nil {
		t.Fatal(err)
	}

	c, _ := tf.NewTensor(float32(3.0))
	outputs, err := sess.Run(
		map[tf.Output]*tf.Tensor{x: c},
		[]tf.Output{grads1[0]},
		nil)
	if err != nil {
		t.Fatal(err)
	}
	if outputs[0].Value().(float32) != 108.0 {
		t.Fatalf("Got %v, wanted float 108.0", outputs[0].Value())
	}
}

func TestValidateGradientsNames(t *testing.T) {
	var (
		s  = NewScope()
		x  = Placeholder(s.SubScope("x"), tf.Float)
		y0 = Square(s.SubScope("y0"), x)
	)

	grads0 := Gradients(s, "", []tf.Output{y0}, []tf.Output{x})
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(grads0[0].Op.Name(), "Gradients/") {
		t.Fatalf("Got name %v, wanted started with Gradients/", grads0[0].Op.Name())
	}

	grads1 := Gradients(s, "", []tf.Output{y0}, []tf.Output{x})
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(grads1[0].Op.Name(), "Gradients_1/") {
		t.Fatalf("Got name %v, wanted started with Gradients_1/", grads1[0].Op.Name())
	}

	grads2 := Gradients(s, "more_gradients", []tf.Output{y0}, []tf.Output{x})
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(grads2[0].Op.Name(), "more_gradients/") {
		t.Fatalf("Got name %v, wanted started with more_gradients/", grads2[0].Op.Name())
	}

	sub := s.SubScope("sub")
	grads3 := Gradients(sub, "even_more_gradients", []tf.Output{y0}, []tf.Output{x})
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(grads3[0].Op.Name(), "sub/even_more_gradients/") {
		t.Fatalf("Got name %v, wanted started with sub/even_more_gradients/", grads3[0].Op.Name())
	}

	grads4 := Gradients(sub, "even_more_gradients", []tf.Output{y0}, []tf.Output{x})
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(grads4[0].Op.Name(), "sub/even_more_gradients_1/") {
		t.Fatalf("Got name %v, wanted started with sub/even_more_gradients_1/", grads4[0].Op.Name())
	}
}
