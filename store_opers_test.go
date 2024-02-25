package ircd

import "testing"

func TestOperatorStore(t *testing.T) {
	os := NewOperatorStore()

	t.Run("add op", func(t *testing.T) {
		os.add("username", "supercomplexpassword")
	})

	t.Run("auth success", func(t *testing.T) {
		ok := os.auth("username", "supercomplexpassword")
		if !ok {
			t.Errorf("auth not successful when it should be")
		}
	})

	t.Run("auth failure", func(t *testing.T) {
		ok := os.auth("username", "notthepassword")
		if ok {
			t.Errorf("auth successful when it should not be")
		}
	})

	t.Run("user does not exist", func(t *testing.T) {
		ok := os.auth("zcxvxcv", "notthepassword")
		if ok {
			t.Errorf("auth with non-existent user successful when it should not be")
		}
	})
}
