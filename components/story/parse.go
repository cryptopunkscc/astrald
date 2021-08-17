package story

func ParseType(story []byte) []byte {
	return story[BeginDynamicData:int(story[BeginTypeSize])]
}

func ParseAuthor(story []byte) []byte {
	start := int(story[BeginTypeSize])
	end := int(story[BeginAuthorSize])
	return story[BeginDynamicData:][start:end]
}
