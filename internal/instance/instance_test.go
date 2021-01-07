package instance

import (
	"testing"

	"github.com/lfmunoz/cobweb/test"
)

func TestConcurrentMap_Simple(t *testing.T) {

	instances := []Instance{
		{1, "nodeId0", "192.168.0.0", make([]string, 0), false},
		{2, "nodeId1", "192.168.0.1", make([]string, 0), false},
		{3, "nodeId2", "192.168.0.2", make([]string, 0), false},
	}

	t.Log("ConcurrentMap should work.")
	{
		t.Logf("\t [Save] ")
		{
			for _, i := range instances {
				Save(i)
			}
			t.Log("\t\t Save should work", test.CheckMark)
		}
		t.Logf("\t [LoadByNodeId] ")
		{
			result, ok := LoadByNodeId(instances[0].NodeId)
			if result.Address != instances[0].Address || !ok {
				t.Errorf("\t\t Failed to read %s %v",
					instances[0].NodeId, test.BallotX)
			} else {
				t.Log("\t\t Load should work", test.CheckMark)
			}
		}
		t.Logf("\t [LoadById] ")
		{
			result, ok := LoadById(1)
			if result.Address != instances[0].Address || !ok {
				t.Errorf("\t\t Failed to read %s %v",
					instances[0].NodeId, test.BallotX)
			} else {
				t.Log("\t\t Load should work", test.CheckMark)
			}
		}
		t.Logf("\t [DeleteById] ")
		{
			Delete(instances[0].Id)
			t.Log("\t\t Delete should work", test.CheckMark)
		}
		t.Logf("\t [Read All] ")
		{
			results := All()
			if len(results) != len(instances)-1 {
				t.Errorf("\t\t Failed to read all elements %d is not %d %v",
					len(results), len(instances)-1, test.BallotX)
			} else {
				t.Log("\t\t Reading all should work", test.CheckMark)

			}
		}
	}
}
