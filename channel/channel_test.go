package channel

import (
	"testing"
)

func TestParse(t *testing.T) {
	channelName := "#dev"

	c, err := New(channelName)
	if err != nil {
		t.Error("channel name is valid")
	}

	if c.Name != "dev" {
		t.Error("Name of this channel should be dev")
	}

	if c.Namespace != NS_NETWORK {
		t.Error("Failed to parse the channel namespace")
	}
}
