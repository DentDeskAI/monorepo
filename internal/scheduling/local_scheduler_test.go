package scheduling

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dentdesk/dentdesk/internal/store"
)

func TestLocalDoctorToScheduler(t *testing.T) {
	specialty := "therapist"

	tests := []struct {
		name   string
		input  store.Doctor
		expect Doctor
	}{
		{
			name: "with specialty",
			input: store.Doctor{
				ID:        uuid.New(),
				Name:      "Dr. A",
				Specialty: &specialty,
				SeqID:     12,
			},
			expect: Doctor{
				ID:          "12",
				Name:        "Dr. A",
				Specialties: []string{"therapist"},
			},
		},
		{
			name: "without specialty",
			input: store.Doctor{
				ID:    uuid.New(),
				Name:  "Dr. B",
				SeqID: 7,
			},
			expect: Doctor{
				ID:          "7",
				Name:        "Dr. B",
				Specialties: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, localDoctorToScheduler(tt.input))
		})
	}
}

func TestLocalPatientToScheduler(t *testing.T) {
	name := "Patient A"

	tests := []struct {
		name   string
		input  store.Patient
		expect Patient
	}{
		{
			name: "with name",
			input: store.Patient{
				ID:    uuid.New(),
				SeqID: 21,
				Phone: "+77001234567",
				Name:  &name,
			},
			expect: Patient{
				ID:        21,
				Name:      "Patient A",
				Number:    "21",
				Phone:     strPtr("+77001234567"),
				IsChild:   false,
				Comment:   "",
				WhereKnow: "",
			},
		},
		{
			name: "without name",
			input: store.Patient{
				ID:    uuid.New(),
				SeqID: 8,
				Phone: "+77007654321",
			},
			expect: Patient{
				ID:        8,
				Name:      "",
				Number:    "8",
				Phone:     strPtr("+77007654321"),
				IsChild:   false,
				Comment:   "",
				WhereKnow: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := localPatientToScheduler(tt.input)
			assert.Equal(t, tt.expect.ID, got.ID)
			assert.Equal(t, tt.expect.Name, got.Name)
			assert.Equal(t, tt.expect.Number, got.Number)
			assert.Equal(t, *tt.expect.Phone, *got.Phone)
			assert.Equal(t, tt.expect.IsChild, got.IsChild)
			assert.Equal(t, tt.expect.Comment, got.Comment)
			assert.Equal(t, tt.expect.WhereKnow, got.WhereKnow)
			assert.Nil(t, got.Gender)
			assert.Nil(t, got.IIN)
			assert.Nil(t, got.Birth)
		})
	}
}

func TestLocalStatusMappings(t *testing.T) {
	assert.Equal(t, 0, localStatusToInt["scheduled"])
	assert.Equal(t, 1, localStatusToInt["confirmed"])
	assert.Equal(t, 2, localStatusToInt["cancelled"])
	assert.Equal(t, 4, localStatusToInt["completed"])
	assert.Equal(t, 2, localStatusToInt["no_show"])

	assert.Equal(t, "scheduled", intToLocalStatus[0])
	assert.Equal(t, "confirmed", intToLocalStatus[1])
	assert.Equal(t, "cancelled", intToLocalStatus[2])
	assert.Equal(t, "completed", intToLocalStatus[3])
	assert.Equal(t, "completed", intToLocalStatus[4])
	assert.Equal(t, "confirmed", intToLocalStatus[5])
	assert.Equal(t, "scheduled", intToLocalStatus[6])
}

func strPtr(s string) *string {
	return &s
}
