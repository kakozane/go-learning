package main

import (
	"reflect"
	"testing"
)

func TestDeleteCourseByID(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ids     []string
		id      string
		want    []string
		deleted bool
	}{
		{"middle", []string{"1", "2", "3", "4"}, "2", []string{"1", "3", "4"}, true},
		{"first", []string{"1", "2"}, "1", []string{"2"}, true},
		{"last", []string{"1", "2"}, "2", []string{"1"}, true},
		{"only", []string{"1"}, "1", []string{}, true},
		{"missing", []string{"1", "2"}, "9", []string{"1", "2"}, false},
		{"empty", nil, "1", []string{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			courses := make([]Course, len(tc.ids))
			for i, id := range tc.ids {
				courses[i] = Course{ID: id, Name: "Course " + id}
			}
			got, deleted := deleteCourseByID(courses, tc.id)
			ids := make([]string, len(got))
			for i, course := range got {
				ids[i] = course.ID
				if course.Name != "Course "+course.ID {
					t.Errorf("unexpected course data: %+v", course)
				}
			}
			if deleted != tc.deleted || !reflect.DeepEqual(ids, tc.want) {
				t.Fatalf("got %v, %v; want %v, %v", ids, deleted, tc.want, tc.deleted)
			}
		})
	}
}

func TestCourseLookupAndRename(t *testing.T) {
	courses := []Course{{ID: "1", Name: "Original"}}
	course, found := findCourseByID(courses, "1")
	if !found || course.Name != "Original" {
		t.Fatal("existing course was not found")
	}
	course.Name = "Copy"
	if courses[0].Name != "Original" {
		t.Fatal("lookup result should be a value copy")
	}
	err, index := updateNameByID(courses, "1", "  Renamed\t")
	if err != nil || index != 0 || courses[0].Name != "Renamed" {
		t.Fatal("rename did not update the original element")
	}
	if _, found := findCourseByID(courses, "missing"); found {
		t.Fatal("missing course was reported as found")
	}
	if err, index := updateNameByID(courses, "missing", "Wrong"); err == nil || err.Error() != "课程不存在" || index != -1 || courses[0].Name != "Renamed" {
		t.Fatal("missing ID should not update any course")
	}
}

func TestRenameRejectsBlankName(t *testing.T) {
	for _, name := range []string{"", " \t\n", "　"} {
		t.Run(name, func(t *testing.T) {
			courses := []Course{{ID: "1", Name: "Original", Completed: true}}
			before := append([]Course(nil), courses...)
			err, index := updateNameByID(courses, "1", name)
			if err == nil || err.Error() != "课程名称不能为空" || index != -1 {
				t.Fatalf("expected blank-name error and -1, got %v and %d", err, index)
			}
			if !reflect.DeepEqual(courses, before) {
				t.Fatal("invalid name must not change course data")
			}
		})
	}
}
