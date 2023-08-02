package database

func StickerTagsUnique(in []*StickerTag) (out []*StickerTag) {
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
