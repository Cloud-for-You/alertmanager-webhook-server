package receivers

import "fmt"

type Receiver interface {
    Producer(data []byte) error
}

// Hlavní receiver, který deleguje na vybranou implementaci
type MainReceiver struct {
    implementation Receiver
}

// NewMainReceiver vrací nový hlavní receiver s konkrétní implementací
func NewMainReceiver(impl Receiver) *MainReceiver {
    return &MainReceiver{implementation: impl}
}

// Přijme data a deleguje zpracování na konkrétní implementaci
func (m *MainReceiver) Receive(data []byte) error {
    if m.implementation == nil {
        return fmt.Errorf("no receiver implementation provided")
    }
    return m.implementation.Producer(data)
}
