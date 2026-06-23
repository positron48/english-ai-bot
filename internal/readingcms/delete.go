package readingcms

import (
	"fmt"
	"os"
)

// DeletePublished removes a text from course and bundle catalogs.
func (s *Service) DeletePublished(courseCode, textID string, removeCMSDraft bool) error {
	course, err := s.paths.Course(courseCode)
	if err != nil {
		return err
	}
	idx, err := loadReadingIndex(course.GrammarDir)
	if err != nil {
		return err
	}
	rel, ok := idx.Texts[textID]
	if !ok {
		return fmt.Errorf("published text not found in course catalog: %s", textID)
	}
	if err := applyReadingTextDeletion(course.GrammarDir, textID, rel); err != nil {
		return err
	}
	if st, err := os.Stat(course.BundleDir); err == nil && st.IsDir() {
		bidx, err := loadReadingIndex(course.BundleDir)
		if err == nil {
			if brel, ok := bidx.Texts[textID]; ok {
				_ = applyReadingTextDeletion(course.BundleDir, textID, brel)
			}
		}
	}
	if removeCMSDraft {
		_ = s.store.DeleteDraft(textID)
	}
	return nil
}

// DeleteDraft removes a CMS draft only.
func (s *Service) DeleteDraft(textID string) error {
	return s.store.DeleteDraft(textID)
}
