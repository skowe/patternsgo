package observer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"
)

type testEvent struct {
	format string
}

type ObserverError struct {
	Val string
}

func (o *ObserverError) Error() string {
	return o.Val
}
func (t *testEvent) Trigger(m Message) error {
	if m.Err() != nil {
		return &ObserverError{
			Val: m.Err().Error(),
		}
	}
	data, err := m.Format(t.format)
	if err != nil {
		return &ObserverError{
			Val: err.Error(),
		}
	}
	log.Println(string(data))
	return nil
}

type testMessage struct {
	Person struct {
		Age  int
		Name string
	}
	err error
}

func (t testMessage) Err() error {
	return t.err
}
func (t testMessage) Format(formatType string) ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0))
	switch formatType {
	case "json":
		encoder := json.NewEncoder(b)
		err := encoder.Encode(t.Person)
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	case "plaintext":
		format := "Age: %d\nName: %s\n"
		return []byte(fmt.Sprintf(format, t.Person.Age, t.Person.Name)), nil
	default:
		return nil, errors.New("No such type defined")
	}
}

type testTarget struct {
	Age  int
	Name string
}

func (t *testTarget) WatchYourself(communicate chan Message, err chan error) {
	curName := t.Name
	curAge := t.Age

	defer func() {
		close(communicate)
		close(err)
	}()

	for i := 0; i < 10; i += 1 {
		if t.Name != curName || t.Age != curAge {
			communicate <- testMessage{
				Person: struct {
					Age  int
					Name string
				}{Age: t.Age, Name: t.Name},
				err: nil,
			}
			curName = t.Name
			curAge = t.Age
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func TestObserver(t *testing.T) {
	b := bytes.NewBuffer(make([]byte, 0))
	curLen := 0
	log.SetOutput(b)
	ev1 := testEvent{
		format: "plaintext",
	}
	ev2 := testEvent{
		format: "json",
	}

	target := &testTarget{
		Age:  17,
		Name: "Alex",
	}
	o := New(target)

	o.AddObserver(&ev1)
	o.AddObserver(&ev2)
	errs := make(chan error, 1)
	go o.Watch(errs)
	time.Sleep(time.Second * 2)
	target.Age = 88
	for err := range errs {
		t.Error(err)
	}
	if curLen == b.Len() {
		t.Error("Failed to log change")
	}

	fmt.Println(b.String())
}

func TestObserveError(t *testing.T) {
	b := bytes.NewBuffer(make([]byte, 0))
	log.SetOutput(b)
	ev1 := testEvent{
		format: "plaintext",
	}
	ev2 := testEvent{
		format: "xml",
	}

	target := &testTarget{
		Age:  17,
		Name: "Alex",
	}
	o := New(target)

	o.AddObserver(&ev1)
	o.AddObserver(&ev2)
	errs := make(chan error, 1)
	go o.Watch(errs)
	time.Sleep(time.Second * 2)
	target.Age = 88
	for err := range errs {
		time.Sleep(time.Millisecond * 40)
		_, ok := <-errs
		if err != nil && !ok {
			fmt.Println(err)
			return
		}
	}
	t.Error("Failed to catch error")
}
