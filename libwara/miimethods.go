package libwara

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Helper methods

// set multiple bits
func setBits(b *Mii, byteindex uint, offset uint8, size uint, val uint64) {
	//we are working on 1 byte, and we do not cross over
	if uint(offset)+size <= 8 {
		mask := ^(byte((0xFF >> (8 - size)) << offset))
		(*b)[byteindex] &= mask
		//temp := (byte(val) >> (8 - size) << offset)
		temp := byte(val) << (offset)
		(*b)[byteindex] |= temp
		return
	}

	scratch := val
	//set bits on first byte
	mask := ^(byte(0xFF >> offset))
	(*b)[byteindex] &= mask
	(*b)[byteindex] |= byte(scratch) >> offset
	scratch >>= (8 - offset)

	//set full bytes
	fullBytes := (size - (8 - uint(offset))) / 8
	for i := uint(0); i < fullBytes; i++ {
		(*b)[byteindex+1+i] = byte(scratch)
		scratch >>= 8
	}

	//set remaining bytes
	remainder := (size - (8 - uint(offset))) % 8
	if remainder != 0 {
		mask = (0xFF >> remainder)
		(*b)[byteindex+fullBytes+1] &= mask
		temp := byte(scratch) << byte(8-remainder)
		(*b)[byteindex+fullBytes+1] |= temp
	}
}

func getBits(b *Mii, byteindex uint, offset uint8, size uint) uint64 {
	//all within same byte
	if uint(offset)+size <= 8 {
		mask := (byte((0xFF << (8 - size)) >> offset))
		temp := uint64(((*b)[byteindex] & mask) << offset)
		return temp
	}

	//get bits from first byte
	mask := ^(byte(0xFF) << (8 - offset))
	ret := uint64(((*b)[byteindex]&mask)<<offset) >> offset

	//get full bytes
	fullBytes := (size - uint(8-offset)) / 8
	for i := uint(0); i < fullBytes; i++ {
		bits := uint64((*b)[byteindex+1+i])
		bits <<= ((i * 8) + uint(8-offset))
		ret |= bits
	}

	//get remaining bytes
	remainder := (size - uint(8-offset)) % 8
	if remainder != 0 {
		bits := uint64((*b)[byteindex+fullBytes+1])
		temp := uint64(^byte((0xFF >> (remainder))))
		bits &= temp
		bits >>= (8 - remainder)
		bits <<= ((fullBytes * 8) + (remainder))
		ret |= bits
	}

	return ret
}

// scale an int from one range to another
// rounds to nearest whole number
func scale(val, loIn, hiIn, loOut, hiOut uint64) uint64 {
	if loOut == hiOut || loIn == hiIn {
		return loOut
	}

	sub := func(a, b uint64) uint64 {
		if a < b {
			return b - a
		} else {
			return a - b
		}
	}

	ret := uint64(math.Round(float64(loOut) + float64(sub(hiOut, loOut))*(float64(sub(val, loIn))/float64(sub(hiIn, loIn)))))

	if ret < loOut {
		ret = loOut
	}
	if ret > hiOut {
		ret = hiOut
	}

	return ret
}

// Adjusts a value to fall within a provided range
func adjustValue(b *Mii, byteindex uint, offset byte, size uint, lo, hi uint64) {
	rawVal := getBits(b, byteindex, offset, size)
	maxVal := ^(^uint64(0x00) << size)
	val := scale(rawVal, 0x00, maxVal, lo, hi)
	setBits(b, byteindex, offset, size, val)
}

// Sets an attribute
// Automatically fixes the CRC
func (m *Mii) setAttribute(attribute int, value uint64) error {
	attributeStruct := MiiFormat[attribute]
	if !(attributeStruct.MinVal <= value && value <= attributeStruct.MaxVal) {
		var msg strings.Builder
		msg.WriteString("expected value between ")
		msg.WriteString(strconv.FormatUint(attributeStruct.MinVal, 10))
		msg.WriteString(" and ")
		msg.WriteString(strconv.FormatUint(attributeStruct.MaxVal, 10))
		msg.WriteString(", got ")
		msg.WriteString(strconv.FormatUint(value, 10))
		return errors.New(msg.String())
	}
	setBits(m, attributeStruct.ByteOffset, attributeStruct.BitOffset, attributeStruct.Size, uint64(value))
	m.FixCRC()
	return nil
}

// Gets the value of an attribute
func (m *Mii) getAttribute(attribute int) uint64 {
	attributeStruct := MiiFormat[attribute]
	value := getBits(m, attributeStruct.ByteOffset, attributeStruct.BitOffset, attributeStruct.Size)
	return value
}

// Sets a flag
// Automatically fixes the CRC
func (m *Mii) setFlag(flag int, value bool) {
	set := 0
	if value {
		set = 1
	}
	attribute := MiiFormat[flag]
	setBits(m, attribute.ByteOffset, attribute.BitOffset, attribute.Size, uint64(set))
	m.FixCRC()
}

// Gets a flag
func (m *Mii) getFlag(flag int) bool {
	attribute := MiiFormat[flag]
	value := getBits(m, attribute.ByteOffset, attribute.BitOffset, attribute.Size)
	return value == 1
}

// Calculates and sets the CRC for a Mii
func (m *Mii) FixCRC() {
	crc := uint16(0x0000)
	for _, currentByte := range m[:94] {
		for bit := 7; bit >= 0; bit-- {
			var flag uint16
			if (crc & 0x8000) != 0 {
				flag = 0x1021
			} else {
				flag = 0
			}
			crc = (((crc << 1) | ((uint16(currentByte) >> bit) & 0x0001)) ^ flag)
		}
	}

	for i := 16; i > 0; i-- {
		var flag uint16
		if (crc & 0x8000) != 0 {
			flag = 0x1021
		} else {
			flag = 0
		}
		crc = ((crc << 1) ^ flag)
	}

	m[94] = byte(crc >> 8)
	m[95] = byte(crc)
}

// Mii modification methods

// Converts a base64 encoded mii to a Mii, or provides an empty Mii if no string is provided
func InitMii(encoded ...string) (*Mii, error) {
	if len(encoded) == 0 {
		return &Mii{}, nil
	}

	mii, err := base64.StdEncoding.DecodeString(encoded[0])
	if err != nil {
		return nil, err
	}

	if len(mii) != 96 {
		return nil, errors.New("string does not encode a mii")
	}

	t := Mii{}
	for i := range mii {
		t[i] = mii[i]
	}

	return &t, nil
}

// Sets the Version attribute of a Mii
func (m *Mii) SetVersion(version uint64) error {
	if version != 0 && version != 3 {
		return errors.New("expected value 0 or 3, not " + strconv.FormatUint(version, 10))
	}
	attribute := MiiFormat[versionAttribute]
	setBits(m, attribute.ByteOffset, attribute.BitOffset, attribute.Size, version)

	m.FixCRC()
	return nil
}

// Gets the Version attribute of a Mii
func (m *Mii) GetVersion() uint64 {
	return m.getAttribute(versionAttribute)
}

// Sets the Allow Copy flag of a Mii
func (m *Mii) SetCopy(allowCopy bool) {
	m.setFlag(copyAttribute, allowCopy)
}

// Gets the Allow Copy flag of a Mii
func (m *Mii) HasCopy() bool {
	return m.getFlag(copyAttribute)
}

// Sets the Profanity flag of a Mii
func (m *Mii) SetProfanity(hasProfanity bool) {
	m.setFlag(profanityAttribute, hasProfanity)
}

// Gets the Profanity flag of a Mii
func (m *Mii) HasProfanity() bool {
	return m.getFlag(profanityAttribute)
}

// Sets the Region Lock attribute of a Mii
func (m *Mii) SetRegionLock(region uint64) error {
	return m.setAttribute(regionLockAttribute, region)
}

// Gets the Region Lock attribute of a Mii
func (m *Mii) GetRegionLock() uint64 {
	return m.getAttribute(regionLockAttribute)
}

// Sets the Character Set attribute of a Mii
func (m *Mii) SetCharSet(charSet uint64) error {
	return m.setAttribute(characterSetAttribute, charSet)
}

// Gets the Character Set attribute of a Mii
func (m *Mii) GetCharSet() uint64 {
	return m.getAttribute(characterSetAttribute)
}

// Sets the 3DS Page attribute of a Mii
func (m *Mii) Set3dsPage(page uint64) error {
	return m.setAttribute(page3dsAttribute, page)
}

// Gets the 3DS Page attribute of a Mii
func (m *Mii) Get3dsPage() uint64 {
	return m.getAttribute(page3dsAttribute)
}

// Sets the 3DS Slot attribute of a Mii
func (m *Mii) Set3dsSlot(slot uint64) error {
	return m.setAttribute(slot3dsAttribute, slot)
}

// Gets the 3DS Slot attribute of a Mii
func (m *Mii) Get3dsSlot() uint64 {
	return m.getAttribute(slot3dsAttribute)
}

// Sets the Devide Origin attribute of a Mii
func (m *Mii) SetDeviceOrigin(origin DeviceOrigin) error {
	err := m.setAttribute(deviceOriginAttribute, uint64(origin))
	if err != nil {
		return err
	}
	if origin == 2 {
		m.setFlag(dsMiiAttribute, true)
	}

	return nil
}

// Gets the Device Origin attribute of a Mii
func (m *Mii) GetDeviceOrigin() uint64 {
	return m.getAttribute(deviceOriginAttribute)
}

// Sets the Device ID attribute of a Mii
func (m *Mii) SetDeviceID(id uint64) error {
	return m.setAttribute(deviceIdAttribute, id)
}

// Gets the Device ID attribute of a Mii
func (m *Mii) GetDeviceID() uint64 {
	return m.getAttribute(deviceIdAttribute)
}

// Sets the Normal Mii flag of a Mii
func (m *Mii) SetNormalMii(isNormalMii bool) {
	m.setFlag(normalMiiAttribute, isNormalMii)
}

// Gets the Normal Mii flag of a Mii
func (m *Mii) IsNormalMii() bool {
	return m.getFlag(normalMiiAttribute)
}

func (m *Mii) SetDSMii(dsMii bool) {
	m.setFlag(dsMiiAttribute, dsMii)
}

// Gets the DS Mii flag of a Mii
func (m *Mii) IsDSMii() bool {
	return m.getFlag(dsMiiAttribute)
}

// Sets the Non-User Mii flag of a Mii
func (m *Mii) SetNonUserMii(isNonUser bool) {
	m.setFlag(nonUserMiiAttribute, isNonUser)
}

// Gets the Non-User Mii flag of a Mii
func (m *Mii) IsNonUserMii() bool {
	return m.getFlag(nonUserMiiAttribute)
}

// Sets the Valid flag of a Mii
func (m *Mii) SetValid(isValid bool) {
	m.setFlag(isValidAttribute, isValid)
}

// Gets the Valid flag of a Mii
func (m *Mii) IsValid() bool {
	return m.getFlag(isValidAttribute)
}

// Sets the Creation Time attribute of a Mii
func (m *Mii) SetCreationTime(creationTime uint64) error {
	return m.setAttribute(creationTimeAttribute, creationTime)
}

// Gets the Creation Time attribute of a Mii
func (m *Mii) GetCreationTime() uint64 {
	return m.getAttribute(creationTimeAttribute)
}

// Sets the Console MAC attribute of a Mii
func (m *Mii) SetConsoleMAC(mac uint64) error {
	return m.setAttribute(systemMacAttribute, mac)
}

// Gets the Console MAC attribute of a Mii
func (m *Mii) GetConsoleMAC() uint64 {
	return m.getAttribute(systemMacAttribute)
}

// Sets the Gender attribute of a Mii
func (m *Mii) SetGender(gender Gender) error {
	return m.setAttribute(genderAttribute, uint64(gender))
}

// Gets the Gender attribute of a Mii
func (m *Mii) GetGender() Gender {
	return Gender(m.getAttribute(genderAttribute))
}

// Sets the Birth Month attribute of a Mii
func (m *Mii) SetBirthMonth(birthMonth uint64) error {
	return m.setAttribute(birthMonthAttribute, birthMonth)
}

// Gets the Birth Month attribute of a Mii
func (m *Mii) GetBirthMonth() uint64 {
	return m.getAttribute(birthMonthAttribute)
}

// Sets the Birth Month attribute of a Mii
func (m *Mii) SetBirthDay(birthDay uint64) error {
	return m.setAttribute(birthDayAttribute, birthDay)
}

// Gets the Birth Month attribute of a Mii
func (m *Mii) GetBirthDay() uint64 {
	return m.getAttribute(birthDayAttribute)
}

// Sets the Favorite Color attribute of a Mii
func (m *Mii) SetFavoriteColor(favoriteColor uint64) error {
	return m.setAttribute(favoriteColorAttribute, favoriteColor)
}

// Gets the Favorite Color attribute of a Mii
func (m *Mii) GetFavoriteColor() uint64 {
	return m.getAttribute(favoriteColorAttribute)
}

// Sets the Favorite flag of a Mii
func (m *Mii) SetFavorite(isFavorite bool) {
	m.setFlag(favoriteAttribute, isFavorite)
}

// Gets the Favorite flag of a Mii
func (m *Mii) IsFavorite() bool {
	return m.getFlag(favoriteAttribute)
}

// Sets the Name attribute of a Mii
func (m *Mii) SetMiiName(miiName string) error {
	if len(miiName) > 10 {
		return errors.New("mii name can have no more than 10 characters, got " + strconv.Itoa(len(miiName)))
	}
	offset := int(MiiFormat[miiNameAttribute].ByteOffset)
	nameBytes := []rune(miiName)
	r := 0
	for i := 0; i < 20; i++ {
		if r >= (len(nameBytes)) {
			break
		}

		if i%2 == 0 {
			m[offset+i] = byte(nameBytes[r])
		} else {
			m[offset+i] = byte((nameBytes[r] & 0xFF00) >> 8)
			r++
		}
	}

	m.FixCRC()

	return nil
}

// Gets the Name attribute of a Mii
func (m *Mii) GetMiiName() string {
	name := []rune{}
	offset := int(MiiFormat[miiNameAttribute].ByteOffset)
	r := 0
	for i := 0; i < 20; i++ {
		if i%2 == 0 {
			name[r] = rune(m[offset+i])
		} else {
			name[r] |= (rune(m[offset+i]) << 8)
			r++
		}
	}

	return string(name)
}

// Sets the Height attribute of a Mii
func (m *Mii) SetHeight(height uint64) error {
	return m.setAttribute(heightAttribute, height)
}

// Gets the Height attribute of a Mii
func (m *Mii) GetHeight() uint64 {
	return m.getAttribute(heightAttribute)
}

// Sets the Build attribute of a Mii
func (m *Mii) SetBuild(build uint64) error {
	return m.setAttribute(buildAttribute, build)
}

// Gets the Build attribute of a Mii
func (m *Mii) GetBuild() uint64 {
	return m.getAttribute(buildAttribute)
}

// Sets the Disable Sharing flag of a Mii
func (m *Mii) SetDisableSharing(disableShare bool) {
	m.setFlag(disableShareAttribute, disableShare)
}

// Gets the Disable Share flag of a Mii
func (m *Mii) HasDisabledSharing() bool {
	return m.getFlag(disableShareAttribute)
}

// Sets the Face Type attribute of a Mii
func (m *Mii) SetFaceType(faceType uint64) error {
	return m.setAttribute(faceTypeAttribute, faceType)
}

// Gets the Face Type attribute of a Mii
func (m *Mii) GetFaceType() uint64 {
	return m.getAttribute(faceTypeAttribute)
}

// Sets the Favorite Color attribute of a Mii
func (m *Mii) SetSkinColor(skinColor uint64) error {
	return m.setAttribute(skinColorAttribute, skinColor)
}

// Gets the Favorite Color attribute of a Mii
func (m *Mii) GetSkinColor() uint64 {
	return m.getAttribute(skinColorAttribute)
}

// Sets the Wrinkle Type attribute of a Mii
func (m *Mii) SetWrinklesType(wrinkleType uint64) error {
	return m.setAttribute(wrinkleTypeAttribute, wrinkleType)
}

// Gets the Wrinkle Type attribute of a Mii
func (m *Mii) GetWrinklesType() uint64 {
	return m.getAttribute(wrinkleTypeAttribute)
}

// Sets the Makeup Type attribute of a Mii
func (m *Mii) SetMakeupType(makeupType uint64) error {
	return m.setAttribute(makeupTypeAttribute, makeupType)
}

// Gets the Makeup Type attribute of a Mii
func (m *Mii) GetMakeupType() uint64 {
	return m.getAttribute(makeupTypeAttribute)
}

// Sets the Hair Type attribute of a Mii
func (m *Mii) SetHairType(hairType uint64) error {
	return m.setAttribute(hairAttribute, hairType)
}

// Gets the Hair Type attribute of a Mii
func (m *Mii) GetHairType() uint64 {
	return m.getAttribute(hairAttribute)
}

// Sets the Hair Color attribute of a Mii
func (m *Mii) SetHairColor(hairColor uint64) error {
	return m.setAttribute(hairColorAttribute, hairColor)
}

// Gets the Hair Color attribute of a Mii
func (m *Mii) GetHairColor() uint64 {
	return m.getAttribute(hairColorAttribute)
}

// Sets the Flip Hair flag of a Mii
func (m *Mii) SetFlippedHair(flipHair bool) {
	m.setFlag(flipHairAttribute, flipHair)
}

// Gets the Flipped Hair flag of a Mii
func (m *Mii) HasFlippedHair() bool {
	return m.getFlag(flipHairAttribute)
}

// Sets the Eye Type attribute of a Mii
func (m *Mii) SetEyeType(eyeType uint64) error {
	return m.setAttribute(eyeTypeAttribute, eyeType)
}

// Gets the Eye Type attribute of a Mii
func (m *Mii) GetEyeType() uint64 {
	return m.getAttribute(eyeTypeAttribute)
}

// Sets the Eye Color attribute of a Mii
func (m *Mii) SetEyeColor(eyeColor uint64) error {
	return m.setAttribute(eyeColorAttribute, eyeColor)
}

// Gets the Eye Color attribute of a Mii
func (m *Mii) GetEyeColor() uint64 {
	return m.getAttribute(eyeColorAttribute)
}

// Sets the Eye Scale attribute of a Mii
func (m *Mii) SetEyeScale(eyeScale uint64) error {
	return m.setAttribute(eyeScaleAttribute, eyeScale)
}

// Gets the Eye Scale attribute of a Mii
func (m *Mii) GetEyeScale() uint64 {
	return m.getAttribute(eyeScaleAttribute)
}

// Sets the Eye Vertical attribute of a Mii
func (m *Mii) SetEyeVertical(eyeVert uint64) error {
	return m.setAttribute(eyeVertAttribute, eyeVert)
}

// Gets the Eye Vertical attribute of a Mii
func (m *Mii) GetEyeVertical() uint64 {
	return m.getAttribute(eyeVertAttribute)
}

// Sets the Eye Rotation attribute of a Mii
func (m *Mii) SetEyeRotation(eyeRot uint64) error {
	return m.setAttribute(eyeRotAttribute, eyeRot)
}

// Gets the Eye Rotation attribute of a Mii
func (m *Mii) GetEyeRotation() uint64 {
	return m.getAttribute(eyeRotAttribute)
}

// Sets the Eye Spacing attribute of a Mii
func (m *Mii) SetEyeSpacing(eyeSpace uint64) error {
	return m.setAttribute(eyeSpaceAttribute, eyeSpace)
}

// Gets the Eye Spacing attribute of a Mii
func (m *Mii) GetEyeSpacing() uint64 {
	return m.getAttribute(eyeSpaceAttribute)
}

// Sets the Eye Y Position attribute of a Mii
func (m *Mii) SetEyeYPosition(eyeYPos uint64) error {
	return m.setAttribute(eyeYPosAttribute, eyeYPos)
}

// Gets the Eye Y Position attribute of a Mii
func (m *Mii) GetEyeYPosition() uint64 {
	return m.getAttribute(eyeYPosAttribute)
}

// Sets the Eyebrow Type attribute of a Mii
func (m *Mii) SetEyebrowType(eyebrowType uint64) error {
	return m.setAttribute(eyebrowTypeAttribute, eyebrowType)
}

// Gets the Eyebrow Type attribute of a Mii
func (m *Mii) GetEyebrowType() uint64 {
	return m.getAttribute(eyebrowTypeAttribute)
}

// Sets the Eyebrow Color attribute of a Mii
func (m *Mii) SetEyebrowColor(eyebrowColor uint64) error {
	return m.setAttribute(eyebrowColorAttribute, eyebrowColor)
}

// Gets the Eyebrow Color attribute of a Mii
func (m *Mii) GetEyebrowColor() uint64 {
	return m.getAttribute(eyebrowColorAttribute)
}

// Sets the Eyebrow Scale attribute of a Mii
func (m *Mii) SetEyebrowScale(eyebrowScale uint64) error {
	return m.setAttribute(eyebrowScaleAttribute, eyebrowScale)
}

// Gets the Eyebrow Scale attribute of a Mii
func (m *Mii) GetEyebrowScale() uint64 {
	return m.getAttribute(eyebrowScaleAttribute)
}

// Sets the Eyebrow Vertical Stretch attribute of a Mii
func (m *Mii) SetEyebrowVertical(eyebrowVert uint64) error {
	return m.setAttribute(eyebrowVertAttribute, eyebrowVert)
}

// Gets the Eyebrow Vertical Stretch attribute of a Mii
func (m *Mii) GetEyebrowVertical() uint64 {
	return m.getAttribute(eyebrowVertAttribute)
}

// Sets the Eyebrow Rotation attribute of a Mii
func (m *Mii) SetEyebrowRotation(eyebrowRot uint64) error {
	return m.setAttribute(eyebrowRotAttribute, eyebrowRot)
}

// Gets the Eyebrow Rotation attribute of a Mii
func (m *Mii) GetEyebrowRotation() uint64 {
	return m.getAttribute(eyebrowRotAttribute)
}

// Sets the Eyebrow Spacing attribute of a Mii
func (m *Mii) SetEyebrowSpacing(eyebrowSpace uint64) error {
	return m.setAttribute(eyebrowSpaceAttribute, eyebrowSpace)
}

// Gets the Eyebrow Type attribute of a Mii
func (m *Mii) GetEyebrowSpacing() uint64 {
	return m.getAttribute(eyebrowSpaceAttribute)
}

// Sets the Eyebrow Type attribute of a Mii
func (m *Mii) SetEyebrowYPosition(eyebrowYPos uint64) error {
	return m.setAttribute(eyebrowYPosAttribute, eyebrowYPos)
}

// Gets the Eyebrow Type attribute of a Mii
func (m *Mii) GetEyebrowYPosition() uint64 {
	return m.getAttribute(eyebrowYPosAttribute)
}

// Sets the Nose Type attribute of a Mii
func (m *Mii) SetNoseType(noseType uint64) error {
	return m.setAttribute(noseTypeAttribute, noseType)
}

// Gets the Nose Type attribute of a Mii
func (m *Mii) GetNoseType() uint64 {
	return m.getAttribute(noseTypeAttribute)
}

// Sets the Nose Scale attribute of a Mii
func (m *Mii) SetNoseScale(noseScale uint64) error {
	return m.setAttribute(noseScaleAttribute, noseScale)
}

// Gets the Nose Scalse attribute of a Mii
func (m *Mii) GetNoseScale() uint64 {
	return m.getAttribute(noseScaleAttribute)
}

// Sets the Nose Y Position attribute of a Mii
func (m *Mii) SetNoseYPosition(noseYPos uint64) error {
	return m.setAttribute(noseYPosAttribute, noseYPos)
}

// Gets the Nose Y Position attribute of a Mii
func (m *Mii) GetNoseYPos() uint64 {
	return m.getAttribute(noseYPosAttribute)
}

// Sets the Mouth Type attribute of a Mii
func (m *Mii) SetMouthType(mouthType uint64) error {
	return m.setAttribute(mouthTypeAttribute, mouthType)
}

// Gets the Mouth Type attribute of a Mii
func (m *Mii) GetMouthType() uint64 {
	return m.getAttribute(mouthTypeAttribute)
}

// Sets the Mouth Color attribute of a Mii
func (m *Mii) SetMouthColor(mouthColor uint64) error {
	return m.setAttribute(mouthColorAttribute, mouthColor)
}

// Gets the Mouth Color attribute of a Mii
func (m *Mii) GetMouthColor() uint64 {
	return m.getAttribute(mouthColorAttribute)
}

// Sets the Mouth Scale attribute of a Mii
func (m *Mii) SetMouthScale(mouthScale uint64) error {
	return m.setAttribute(mouthScaleAttribute, mouthScale)
}

// Gets the Mouth Scale attribute of a Mii
func (m *Mii) GetMouthScale() uint64 {
	return m.getAttribute(mouthScaleAttribute)
}

// Sets the Mouth Stretch attribute of a Mii
func (m *Mii) SetMouthStretch(mouthStretch uint64) error {
	return m.setAttribute(mouthHorPosAttribute, mouthStretch)
}

// Gets the Eyebrow Type attribute of a Mii
func (m *Mii) GetMouthStretch() uint64 {
	return m.getAttribute(mouthHorPosAttribute)
}

// Sets the Mouth Y Position attribute of a Mii
func (m *Mii) SetMouthYPosition(mouthYPos uint64) error {
	return m.setAttribute(mouthYPosAttribute, mouthYPos)
}

// Gets the Mouth Y Position attribute of a Mii
func (m *Mii) GetMouthYPosition() uint64 {
	return m.getAttribute(mouthYPosAttribute)
}

// Sets the Mustache Type attribute of a Mii
func (m *Mii) SetMustacheType(mustacheType uint64) error {
	return m.setAttribute(mustacheTypeAttribute, mustacheType)
}

// Gets the Mustache Type attribute of a Mii
func (m *Mii) GetMustacheType() uint64 {
	return m.getAttribute(mustacheTypeAttribute)
}

// Sets the Beard Type attribute of a Mii
func (m *Mii) SetBeardType(beardType uint64) error {
	return m.setAttribute(beardTypeAttribute, beardType)
}

// Gets the Beard Type attribute of a Mii
func (m *Mii) GetBeardType() uint64 {
	return m.getAttribute(beardTypeAttribute)
}

// Sets the Mustache Scale attribute of a Mii
func (m *Mii) SetMustacheScale(mustacheScale uint64) error {
	return m.setAttribute(mustacheScaleAttribute, mustacheScale)
}

// Gets the Mustache Scale attribute of a Mii
func (m *Mii) GetMustacheScale() uint64 {
	return m.getAttribute(mustacheScaleAttribute)
}

// Sets the Mustache Y Position attribute of a Mii
func (m *Mii) SetMustacheYPosition(mustacheYPos uint64) error {
	return m.setAttribute(mustacheYPosAttribute, mustacheYPos)
}

// Gets the Mustache Y Position attribute of a Mii
func (m *Mii) GetMustacheYPosition() uint64 {
	return m.getAttribute(mustacheYPosAttribute)
}

// Sets the Glasses Type attribute of a Mii
func (m *Mii) SetGlassesType(glassesType uint64) error {
	return m.setAttribute(glassesTypeAttribute, glassesType)
}

// Gets the Glasses Type attribute of a Mii
func (m *Mii) GetGlassesType() uint64 {
	return m.getAttribute(glassesTypeAttribute)
}

// Sets the Glasses Color attribute of a Mii
func (m *Mii) SetGlassesColor(glassesColor uint64) error {
	return m.setAttribute(glassesColorAttribute, glassesColor)
}

// Gets the Glasses Color attribute of a Mii
func (m *Mii) GetGlassesColor() uint64 {
	return m.getAttribute(glassesColorAttribute)
}

// Sets the Glasses Scale attribute of a Mii
func (m *Mii) SetGlassesScale(glassesScale uint64) error {
	return m.setAttribute(glassesScaleAttribute, glassesScale)
}

// Gets the Glasses Scale attribute of a Mii
func (m *Mii) GetGlassesScale() uint64 {
	return m.getAttribute(glassesScaleAttribute)
}

// Sets the Glasses Y Position attribute of a Mii
func (m *Mii) SetGlassesYPosition(glassesYPos uint64) error {
	return m.setAttribute(glassesYPosAttribute, glassesYPos)
}

// Gets the Glasses Y Position attribute of a Mii
func (m *Mii) GetGlassesYPosition() uint64 {
	return m.getAttribute(glassesYPosAttribute)
}

// Sets the Mole Enabled flag of a Mii
func (m *Mii) SetMoleEnabled(moleEnabled bool) {
	m.setFlag(moleEnabledAttribute, moleEnabled)
}

// Gets the Mole Enabled flag of a Mii
func (m *Mii) GetMoleEnabled() bool {
	return m.getFlag(moleEnabledAttribute)
}

// Sets the Mole Scale attribute of a Mii
func (m *Mii) SetMoleScale(moleScale uint64) error {
	return m.setAttribute(moleScaleAttribute, moleScale)
}

// Gets the Mole Scale attribute of a Mii
func (m *Mii) GetMoleScale() uint64 {
	return m.getAttribute(moleScaleAttribute)
}

// Sets the Mole X Position attribute of a Mii
func (m *Mii) SetMoleXPosition(moleXPos uint64) error {
	return m.setAttribute(moleXPosAttribute, moleXPos)
}

// Gets the Mole X Position attribute of a Mii
func (m *Mii) GetMoleXPosition() uint64 {
	return m.getAttribute(moleXPosAttribute)
}

// Sets the Mole Y Position attribute of a Mii
func (m *Mii) SetMoleYPosition(moleYPos uint64) error {
	return m.setAttribute(moleYPosAttribute, moleYPos)
}

// Gets the Mole Y Position attribute of a Mii
func (m *Mii) GetMoleYPosition() uint64 {
	return m.getAttribute(moleYPosAttribute)
}

// Sets the Creator Name attribute of a Mii
func (m *Mii) SetCreatorName(creatorName string) error {
	if len(creatorName) > 10 {
		return errors.New("mii name can have no more than 10 characters, got " + strconv.Itoa(len(creatorName)))
	}
	offset := int(MiiFormat[creatorNameAttribute].ByteOffset)
	nameBytes := []rune(creatorName)
	r := 0
	for i := 0; i < 20; i++ {
		if r >= (len(nameBytes)) {
			break
		}

		if i%2 == 0 {
			m[offset+i] = byte(nameBytes[r])
		} else {
			m[offset+i] = byte((nameBytes[r] & 0xFF00) >> 8)
			r++
		}
	}

	m.FixCRC()

	return nil
}

// Gets the Name attribute of a Mii
func (m *Mii) GetCreatorName() string {
	name := []rune{}
	offset := int(MiiFormat[creatorNameAttribute].ByteOffset)
	r := 0
	for i := 0; i < 20; i++ {
		if i%2 == 0 {
			name[r] = rune(m[offset+i])
		} else {
			name[r] |= (rune(m[offset+i]) << 8)
			r++
		}
	}

	return string(name)
}

// Encodes a Mii to a string
func (m *Mii) Encode() string {
	return base64.StdEncoding.EncodeToString(m[:])
}

// Creates a mii with a specified Mii Name and Creator Name when provided with a seed (string)
func CreateRandomMii(seed, miiName, creatorName string) (*Mii, error) {
	hashData := md5.New().Sum([]byte(seed))

	cycle := createBitCycle(uint64(len(hashData)*8), hashData)

	miiData, err := InitMii()
	if err != nil {
		return nil, err
	}

	skipAttributes := []int{
		versionAttribute,
		copyAttribute,
		profanityAttribute,
		deviceOriginAttribute,
		regionLockAttribute,
		characterSetAttribute,
		blank1Attribute,
		unknown1Attribute,
		blank2Attribute,
		normalMiiAttribute,
		dsMiiAttribute,
		nonUserMiiAttribute,
		isValidAttribute,
		creationTimeAttribute,
		blank3Attribute,
		blank4Attribute,
		miiNameAttribute,
		disableShareAttribute,
		blank5Attribute,
		blank6Attribute,
		blank7Attribute,
		blank8Attribute,
		blank9Attribute,
		blank10Attribute,
		unknown2Attribute,
		blank11Attribute,
		blank12Attribute,
		creatorNameAttribute,
		blank13Attribute,
	}
	for i := 0; i < blank13Attribute+1; i++ {
		skip := false
		for _, j := range skipAttributes {
			if i == j {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		a := MiiFormat[i]
		rawVal := cycle.readBitCycle(uint8(a.Size))
		maxRawVal := ^(^uint64(0x00) << a.Size)
		val := scale(rawVal, 0x00, maxRawVal, a.MinVal, a.MaxVal)

		if err = miiData.setAttribute(i, val); err != nil {
			fmt.Println(i)
			return nil, err
		}
	}

	if err = miiData.SetVersion(0); err != nil {
		return nil, err
	}
	miiData.SetCopy(true)
	miiData.SetProfanity(false)
	if err = miiData.SetRegionLock(0); err != nil {
		return nil, err
	}
	if err = miiData.SetDeviceOrigin(DeviceWiiU); err != nil {
		return nil, err
	}
	if err = miiData.SetCharSet(0); err != nil {
		return nil, err
	}
	miiData.SetNormalMii(true)
	miiData.SetNonUserMii(false)
	miiData.SetValid(true)
	if err = miiData.SetMiiName(miiName); err != nil {
		return nil, err
	}
	if err = miiData.SetCreationTime(0x1000000); err != nil {
		return nil, err
	}
	miiData.SetDisableSharing(false)
	if err = miiData.SetCreatorName(creatorName); err != nil {
		return nil, err
	}
	miiData.SetDSMii(true)

	if err = miiData.SetConsoleMAC(0x40F407A385D6); err != nil {
		return nil, err
	}

	return miiData, nil
}
