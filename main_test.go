package main

import "testing"

func TestCounterResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		counter int
		want    string
	}{
		{1, "counter: 1"},
		{9, "counter: 9"},
	}

	for _, test := range tests {
		test := test

		t.Run(test.want, func(t *testing.T) {
			t.Parallel()

			got := string(CounterResponse(test.counter))

			if test.want != got {
				t.Errorf(`want "%s", got "%s"`, test.want, got)
			}
		})
	}
}
