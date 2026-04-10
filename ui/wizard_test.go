package ui

import (
	"errors"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// mockScreen is a minimal WizardScreen for testing the controller.
type mockScreen struct {
	enterCount int
	leaveCount int
	leaveErr   error
	label      string
}

func (m *mockScreen) Content() fyne.CanvasObject {
	return widget.NewLabel(m.label)
}

func (m *mockScreen) OnEnter(_ *WizardState) {
	m.enterCount++
}

func (m *mockScreen) OnLeave(_ *WizardState) error {
	m.leaveCount++
	return m.leaveErr
}

func newTestController(screens ...*mockScreen) (*WizardController, fyne.Window) {
	w := test.NewWindow(widget.NewLabel("init"))
	state := &WizardState{}
	ifaces := make([]WizardScreen, len(screens))
	for i, s := range screens {
		ifaces[i] = s
	}
	ctrl := NewWizardController(w, state, ifaces, fyne.NewSize(500, 300))
	return ctrl, w
}

func TestShow_CallsOnEnterOnFirstScreen(t *testing.T) {
	s0 := &mockScreen{label: "screen0"}
	ctrl, _ := newTestController(s0)

	ctrl.Show()

	if s0.enterCount != 1 {
		t.Fatalf("expected OnEnter called once, got %d", s0.enterCount)
	}
}

func TestNext_AdvancesAndCallsLifecycle(t *testing.T) {
	s0 := &mockScreen{label: "screen0"}
	s1 := &mockScreen{label: "screen1"}
	ctrl, _ := newTestController(s0, s1)
	ctrl.Show()

	ctrl.Next()

	if s0.leaveCount != 1 {
		t.Fatalf("expected OnLeave on screen0 once, got %d", s0.leaveCount)
	}
	if s1.enterCount != 1 {
		t.Fatalf("expected OnEnter on screen1 once, got %d", s1.enterCount)
	}
}

func TestNext_StopsOnValidationError(t *testing.T) {
	s0 := &mockScreen{label: "screen0", leaveErr: errors.New("invalid")}
	s1 := &mockScreen{label: "screen1"}
	ctrl, _ := newTestController(s0, s1)
	ctrl.Show()

	ctrl.Next()

	if s1.enterCount != 0 {
		t.Fatalf("screen1 OnEnter should not be called when validation fails")
	}
}

func TestNext_DoesNothingOnLastScreen(t *testing.T) {
	s0 := &mockScreen{label: "screen0"}
	ctrl, _ := newTestController(s0)
	ctrl.Show()

	ctrl.Next() // should be a no-op

	if s0.leaveCount != 0 {
		t.Fatalf("OnLeave should not be called on the last screen when calling Next")
	}
}

func TestBack_DecrementsAndCallsLifecycle(t *testing.T) {
	s0 := &mockScreen{label: "screen0"}
	s1 := &mockScreen{label: "screen1"}
	ctrl, _ := newTestController(s0, s1)
	ctrl.Show()
	ctrl.Next() // move to s1

	ctrl.Back()

	if s1.leaveCount != 1 {
		t.Fatalf("expected OnLeave on screen1 once, got %d", s1.leaveCount)
	}
	// s0 OnEnter: once from Show(), once from Back()
	if s0.enterCount != 2 {
		t.Fatalf("expected OnEnter on screen0 twice, got %d", s0.enterCount)
	}
}

func TestBack_IgnoresOnLeaveError(t *testing.T) {
	s0 := &mockScreen{label: "screen0"}
	s1 := &mockScreen{label: "screen1", leaveErr: errors.New("fail")}
	ctrl, _ := newTestController(s0, s1)
	ctrl.Show()
	ctrl.Next() // move to s1

	ctrl.Back() // should still navigate back despite error

	// s0 OnEnter: once from Show(), once from Back()
	if s0.enterCount != 2 {
		t.Fatalf("expected Back to navigate despite OnLeave error, got enterCount %d", s0.enterCount)
	}
}

func TestBack_DoesNothingOnFirstScreen(t *testing.T) {
	s0 := &mockScreen{label: "screen0"}
	ctrl, _ := newTestController(s0)
	ctrl.Show()

	ctrl.Back() // should be a no-op

	if s0.leaveCount != 0 {
		t.Fatalf("OnLeave should not be called on the first screen when calling Back")
	}
}

func TestShow_NoScreensDoesNotPanic(t *testing.T) {
	w := test.NewWindow(widget.NewLabel("init"))
	state := &WizardState{}
	ctrl := NewWizardController(w, state, nil, fyne.NewSize(500, 300))

	// Should not panic.
	ctrl.Show()
}
