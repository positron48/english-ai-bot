package repository

import "fmt"

// TheoryBlockInfo stores normalized references for one theory block.
type TheoryBlockInfo struct {
	ChapterID string
	BlockID   string
	Title     string
	ConceptID string
}

// TheoryBlockIndex is a runtime index built from grammar bundle chapters.
type TheoryBlockIndex struct {
	ByBlockID   map[string]*TheoryBlockInfo
	ByChapterID map[string][]*TheoryBlockInfo
}

// BuildTheoryBlockIndex creates an index of all blocks with type="theory".
func BuildTheoryBlockIndex(contentRepo *GrammarContentRepository) (*TheoryBlockIndex, error) {
	if contentRepo == nil {
		return nil, fmt.Errorf("content repository is nil")
	}
	chapterIDs, err := contentRepo.GetAllChapterIDs()
	if err != nil {
		return nil, fmt.Errorf("get chapter ids: %w", err)
	}

	idx := &TheoryBlockIndex{
		ByBlockID:   make(map[string]*TheoryBlockInfo),
		ByChapterID: make(map[string][]*TheoryBlockInfo),
	}

	for _, chapterID := range chapterIDs {
		chapter, err := contentRepo.GetChapter(chapterID)
		if err != nil {
			return nil, fmt.Errorf("get chapter %s: %w", chapterID, err)
		}
		for _, rawBlock := range chapter.Blocks {
			block, ok := rawBlock.(map[string]interface{})
			if !ok {
				continue
			}
			tp, _ := block["type"].(string)
			if tp != "theory" {
				continue
			}
			blockID, _ := block["id"].(string)
			if blockID == "" {
				continue
			}
			title, _ := block["title"].(string)
			conceptID := ""
			if theory, ok := block["theory"].(map[string]interface{}); ok {
				conceptID, _ = theory["concept_id"].(string)
			}
			info := &TheoryBlockInfo{
				ChapterID: chapterID,
				BlockID:   blockID,
				Title:     title,
				ConceptID: conceptID,
			}
			idx.ByBlockID[blockID] = info
			idx.ByChapterID[chapterID] = append(idx.ByChapterID[chapterID], info)
		}
	}
	return idx, nil
}

