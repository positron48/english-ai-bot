package service

// EnsureVerbFormUserCardsForWord prepares only the newly added vocabulary word.
// Full vocabulary reconciliation remains available for scope/content changes.
func (s *VerbTrainingService) EnsureVerbFormUserCardsForWord(userID, wordCardID int64, scopes []string) error {
	if !s.Enabled() {
		return nil
	}
	linked, err := s.repo.LinkSpanishVerbForWord(wordCardID)
	if err != nil || !linked {
		return err
	}
	rows, err := s.repo.GetLinkedVerbFormsForWord(userID, wordCardID, scopes)
	if err != nil {
		return err
	}
	if err := s.ensureTrainingCardsForRows(rows); err != nil {
		return err
	}
	return s.repo.EnsureUserVerbCardsForWord(userID, wordCardID, scopes)
}
