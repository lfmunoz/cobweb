package instance

import (
	"encoding/json"
	"testing"

	"github.com/lfmunoz/cobweb/test"
)

func TestConcurrentMap_Simple(t *testing.T) {

	instances := []Instance{
		{Id: 1, NodeId: "nodeId0", Address: "192.168.0.0"},
		{Id: 2, NodeId: "nodeId1", Address: "192.168.0.1"},
		{Id: 3, NodeId: "nodeId2", Address: "192.168.0.2"},
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
			DeleteById(instances[0].Id)
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

func TestSerializeDeserialize(t *testing.T) {
	// t.Skip()
	t.Log("Serialize and Deserialize should work.")
	{
		t.Logf("\t [Deserialize Single] ")
		{
			instance1 := []byte(`{"dependencies":[],"local":[{"address":"0.0.0.0","name":"nginx_local","port":8080}],"name":"web0_0","private_ip":"172.17.0.4","public_ip":"172.17.0.4","remote":[{"address":"localhost","name":"nginx_remote","port":80}]}`)
			var m Infrastructure
			err := json.Unmarshal(instance1, &m)
			if err != nil && m.Name != "web_0" {
				t.Error("\t\t Deserialize failed", err, test.BallotX)
			} else {
				t.Log("\t\t Deserialize works", test.CheckMark)
			}
		}
		t.Logf("\t [Deserialize Array] ")
		{
			cluster := []byte(`[{"dependencies":[],"gateway":"172.17.0.1","local":[{"address":"0.0.0.0","name":"nginx_local","port":8080}],"name":"web0_0","private_ip":"172.17.0.4","public_ip":"172.17.0.4","remote":[{"address":"localhost","name":"nginx_remote","port":80}],"scripts":["start_envoy","docker/nginx"]},{"dependencies":[],"gateway":"172.17.0.1","local":[{"address":"0.0.0.0","name":"nginx_local","port":8080}],"name":"web1_0","private_ip":"172.17.0.5","public_ip":"172.17.0.5","remote":[{"address":"localhost","name":"nginx_remote","port":80}],"scripts":["start_envoy","docker/nginx"]},{"dependencies":[],"gateway":"172.17.0.1","local":[{"address":"0.0.0.0","name":"nginx_local","port":8080}],"name":"web1_1","private_ip":"172.17.0.3","public_ip":"172.17.0.3","remote":[{"address":"localhost","name":"nginx_remote","port":80}],"scripts":["start_envoy","docker/nginx"]}]`)
			var m []Infrastructure
			err := json.Unmarshal(cluster, &m)
			if err != nil && m[0].Name != "web_0" {
				t.Error("\t\t Deserialize failed", err, test.BallotX)
			} else {

				t.Log("\t\t Deserialize works", test.CheckMark)
			}

		}
	}
}
