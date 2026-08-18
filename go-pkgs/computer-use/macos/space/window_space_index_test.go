package space

import (
	"reflect"
	"testing"
)

func TestListUserSpacesDenseType0(t *testing.T) {
	got, err := ListUserSpaces(
		WithPlatformGOOS("darwin"),
		WithManagedDisplays([]DisplaySpaces{{
			Spaces: []SpaceInfo{
				{ID: 10, Type: 0, UUID: "aaa"},
				{ID: 99, Type: 4, UUID: "skip"},
				{ID: 11, Type: 0, UUID: "bbb"},
			},
			Current: SpaceInfo{ID: 11, Type: 0, UUID: "bbb"},
		}}),
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []UserSpace{
		{Index: 0, ID: 10, UUID: "aaa"},
		{Index: 1, ID: 11, UUID: "bbb", Current: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestListUserSpacesUnsupportedPlatform(t *testing.T) {
	_, err := ListUserSpaces(WithPlatformGOOS("linux"))
	if err != ErrUnsupportedPlatform {
		t.Fatalf("err=%v want ErrUnsupportedPlatform", err)
	}
}
