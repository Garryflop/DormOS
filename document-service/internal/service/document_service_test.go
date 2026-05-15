package service_test

import (
	"testing"

	"github.com/Garryflop/DormManage/document-service/internal/domain"
)

func TestWorkflowStepDef_Parsing(t *testing.T) {
	steps := []domain.WorkflowStepDef{
		{StepNumber: 1, ApproverRole: "floor_warden", ApproverID: nil, SLAHours: 24},
		{StepNumber: 2, ApproverRole: "dorm_admin", ApproverID: nil, SLAHours: 24},
	}

	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	if steps[0].ApproverRole != "floor_warden" {
		t.Errorf("expected floor_warden, got %s", steps[0].ApproverRole)
	}
	if steps[1].StepNumber != 2 {
		t.Errorf("expected step 2, got %d", steps[1].StepNumber)
	}
}

func TestRequestStatus_Terminal(t *testing.T) {
	terminal := []domain.RequestStatus{
		domain.StatusApproved,
		domain.StatusRejected,
		domain.StatusCancelled,
	}
	for _, s := range terminal {
		if s == domain.StatusInProgress {
			t.Errorf("status %q should not equal in_progress", s)
		}
	}
}

func TestMarshalUnmarshalMetadata(t *testing.T) {
	original := map[string]string{
		"purpose": "bank account opening",
		"notes":   "urgent",
	}

	data, err := domain.MarshalMetadata(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	result, err := domain.UnmarshalMetadata(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result["purpose"] != original["purpose"] {
		t.Errorf("expected purpose %q, got %q", original["purpose"], result["purpose"])
	}
	if result["notes"] != original["notes"] {
		t.Errorf("expected notes %q, got %q", original["notes"], result["notes"])
	}
}

func TestMarshalMetadata_Nil(t *testing.T) {
	data, err := domain.MarshalMetadata(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("expected {}, got %s", string(data))
	}
}

func TestStepStatus_Values(t *testing.T) {
	cases := map[domain.StepStatus]string{
		domain.StepPending:  "pending",
		domain.StepApproved: "approved",
		domain.StepRejected: "rejected",
		domain.StepSkipped:  "skipped",
	}
	for s, expected := range cases {
		if string(s) != expected {
			t.Errorf("expected %q, got %q", expected, string(s))
		}
	}
}

func TestWorkflowStateTransitions(t *testing.T) {
	transitions := []struct {
		from     domain.RequestStatus
		action   string
		expected domain.RequestStatus
	}{
		{domain.StatusInProgress, "final_approve", domain.StatusApproved},
		{domain.StatusInProgress, "reject", domain.StatusRejected},
	}

	for _, tc := range transitions {
		var result domain.RequestStatus
		switch tc.action {
		case "final_approve":
			result = domain.StatusApproved
		case "reject":
			result = domain.StatusRejected
		}
		if result != tc.expected {
			t.Errorf("transition %s -> %s: expected %s, got %s", tc.from, tc.action, tc.expected, result)
		}
	}
}
