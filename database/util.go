package database

func stickersUnique(in []*StickerTag) (out []*StickerTag) {
	out = make([]*StickerTag, 0, len(in))
	checkUnique := make(map[uint64]bool)
	for _, st := range in {
		if !checkUnique[st.StickerID] {
			out = append(out, st)
			checkUnique[st.StickerID] = true
		}
	}
	return
}

func stickersSort(in []*StickerTag, order *StickerOrder) error {
	if err := order.SortStickers(in); err != nil {
		return err
	}
	// immediately write new permutation just in case
	if err := order.UpdateFromStickers(in); err != nil {
		return err
	}
	return nil
}

func uniqueAndSorted(in []*StickerTag, order *StickerOrder) ([]*StickerTag, error) {
	unique := stickersUnique(in)
	if err := stickersSort(unique, order); err != nil {
		return nil, err
	}
	return unique, nil
}
