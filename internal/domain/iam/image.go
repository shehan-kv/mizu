package iam

type Image struct {
	name     ImageName
	mimeType MimeType
}

func NewImage(name ImageName, mimeType MimeType) Image {
	return Image{
		name:     name,
		mimeType: mimeType,
	}
}

func (im Image) Name() ImageName {
	return im.name
}

func (im Image) MimeType() MimeType {
	return im.mimeType
}
