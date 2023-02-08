package libwara

type Mii [96]byte

// Elements that make up a mii attribute
type miiAttribute struct {
	ByteOffset uint   // Number of first byte value appears in
	BitOffset  byte   // Number of first bit within the byte the value appears in
	Size       uint   // How many bits the value takes up
	MinVal     uint64 // Minimum value the attribute can have
	MaxVal     uint64 // Maximum value the attribute can have
}

// List of attributes that make up a mii
var MiiFormat = []miiAttribute{
	{0, 0, 8, 0x03, 0x03},         //version
	{1, 0, 1, 1, 1},               //copy
	{1, 1, 1, 0, 0},               //profanity
	{1, 2, 2, 0, 3},               //region lock
	{1, 4, 2, 0, 3},               //char set
	{1, 6, 2, 0, 0},               //blank 1
	{2, 0, 4, 0, 9},               //page 3ds
	{2, 4, 4, 0, 9},               //slot 3ds
	{3, 0, 4, 0, 0},               //unknown 1
	{3, 4, 3, 1, 4},               //device origin
	{3, 7, 1, 0, 0},               //blank 2
	{4, 0, 64, 0, 0xFFFFFFFFFFFF}, //system MAC

	//swap endian
	{12, 7, 1, 0, 1},          //normal mii
	{12, 6, 1, 0, 1},          //ds mii
	{12, 5, 1, 0, 1},          //non user mii
	{12, 4, 1, 0, 1},          //is valid
	{13, 0, 28, 0, 0xFFFFFFF}, //creation time
	//swap endian

	{16, 0, 48, 0, 0xFFFFFFFFFFFF}, //device id
	{22, 0, 16, 0, 0},              //blank 3
	{24, 0, 1, 0, 1},               //gender
	{25, 1, 4, 1, 12},              //birth month
	{24, 5, 5, 1, 31},              //birth day
	{25, 2, 4, 0, 11},              //favorite color
	{25, 6, 1, 0, 1},               //favorite
	{25, 7, 1, 0, 0},               //blank 4
	{26, 0, 160, 0, 0},             //mii name
	{46, 0, 8, 0, 127},             //height
	{47, 0, 8, 0, 127},             //build
	{48, 0, 1, 0, 1},               //disable share
	{48, 1, 4, 0, 11},              //face type
	{48, 5, 3, 0, 6},               //skin color
	{49, 0, 4, 0, 11},              //wrinkle type
	{49, 4, 4, 0, 11},              //makeup type
	{50, 0, 8, 0, 131},             //hair
	{51, 0, 3, 0, 7},               //hair color
	{51, 3, 1, 0, 1},               //flip hair
	{51, 4, 4, 0, 0},               //blank 5
	{52, 0, 6, 0, 59},              //eye type
	{52, 6, 3, 0, 5},               //eye color
	{53, 1, 4, 0, 7},               //eye scale
	{53, 5, 3, 0, 6},               //eye vert
	{54, 0, 5, 0, 7},               //eye rot
	{54, 5, 4, 0, 12},              //eye space
	{55, 1, 5, 0, 18},              //eye y pos
	{55, 6, 2, 0, 0},               //blank 6
	{56, 0, 5, 0, 24},              //eyebrow type
	{56, 5, 3, 0, 7},               //eyebrow color
	{57, 0, 4, 0, 8},               //eyebrow scale
	{57, 4, 3, 0, 6},               //eyebrow vert
	{57, 7, 1, 0, 0},               //blank 7
	{58, 0, 4, 0, 11},              //eyebrow rot
	{58, 4, 1, 0, 0},               //blank 8
	{58, 5, 4, 0, 12},              //eyebrow space
	{59, 1, 5, 3, 18},              //eyebrow y pos
	{59, 6, 2, 0, 0},               //blank 9
	{60, 0, 5, 0, 17},              //nose type
	{60, 5, 4, 0, 8},               //nose scale
	{61, 1, 5, 0, 18},              //nose y pos
	{61, 6, 2, 0, 0},               //blank 10
	{62, 0, 6, 0, 35},              //mouth type
	{62, 6, 3, 0, 4},               //mouth color
	{63, 1, 4, 0, 8},               //mouth scale
	{63, 5, 3, 0, 6},               //mouth hor
	{64, 0, 5, 0, 18},              //mouth y pos
	{64, 5, 3, 0, 5},               //mustache type
	{65, 0, 8, 0, 0},               //unknown 2
	{66, 0, 3, 0, 6},               //beard type
	{66, 4, 3, 0, 7},               //face hair color
	{66, 6, 4, 0, 8},               //mustache scale
	{67, 2, 5, 0, 16},              //mustache y pos
	{67, 7, 1, 0, 0},               //blank 11
	{68, 0, 4, 0, 8},               //glasses type
	{68, 4, 3, 0, 5},               //glasses color
	{68, 7, 4, 0, 7},               //glasses scale
	{69, 3, 5, 0, 20},              //glasses y pos
	{70, 0, 1, 0, 1},               //mole enabled
	{70, 1, 4, 0, 8},               //mole scale
	{70, 5, 5, 0, 16},              //mole x pos
	{71, 2, 5, 0, 30},              //mole y pos
	{71, 7, 1, 0, 0},               //blank 12
	{72, 0, 160, 0, 0},             //creator name
	{92, 0, 16, 0, 0},              //blank 13
}

// Indexes of attributes in the MiiFormat list
const (
	versionAttribute int = iota
	copyAttribute
	profanityAttribute
	regionLockAttribute
	characterSetAttribute
	blank1Attribute
	page3dsAttribute
	slot3dsAttribute
	unknown1Attribute
	deviceOriginAttribute
	blank2Attribute
	systemMacAttribute

	normalMiiAttribute
	dsMiiAttribute
	nonUserMiiAttribute
	isValidAttribute
	creationTimeAttribute

	deviceIdAttribute
	blank3Attribute
	genderAttribute
	birthMonthAttribute
	birthDayAttribute
	favoriteColorAttribute
	favoriteAttribute
	blank4Attribute
	miiNameAttribute
	heightAttribute
	buildAttribute
	disableShareAttribute
	faceTypeAttribute
	skinColorAttribute
	wrinkleTypeAttribute
	makeupTypeAttribute
	hairAttribute
	hairColorAttribute
	flipHairAttribute
	blank5Attribute
	eyeTypeAttribute
	eyeColorAttribute
	eyeScaleAttribute
	eyeVertAttribute
	eyeRotAttribute
	eyeSpaceAttribute
	eyeYPosAttribute
	blank6Attribute
	eyebrowTypeAttribute
	eyebrowColorAttribute
	eyebrowScaleAttribute
	eyebrowVertAttribute
	blank7Attribute
	eyebrowRotAttribute
	blank8Attribute
	eyebrowSpaceAttribute
	eyebrowYPosAttribute
	blank9Attribute
	noseTypeAttribute
	noseScaleAttribute
	noseYPosAttribute
	blank10Attribute
	mouthTypeAttribute
	mouthColorAttribute
	mouthScaleAttribute
	mouthHorPosAttribute
	mouthYPosAttribute
	mustacheTypeAttribute
	unknown2Attribute
	beardTypeAttribute
	faceHairColorAttribute
	mustacheScaleAttribute
	mustacheYPosAttribute
	blank11Attribute
	glassesTypeAttribute
	glassesColorAttribute
	glassesScaleAttribute
	glassesYPosAttribute
	moleEnabledAttribute
	moleScaleAttribute
	moleXPosAttribute
	moleYPosAttribute
	blank12Attribute
	creatorNameAttribute
	blank13Attribute
)

// Favorite Color values
type FavoriteColor uint64

const (
	ColorRed FavoriteColor = iota
	ColorOrange
	ColorYellow
	ColorYellowGreen
	ColorGreen
	ColorBlue
	ColorSkyBlue
	ColorPink
	ColorPurple
	ColorBrown
	ColorWhite
	ColorBlack
)

// Device Origin values
type DeviceOrigin uint64

const (
	DeviceWii DeviceOrigin = iota + 1
	DeviceDS
	Device3DS
	DeviceWiiU
)

// Gender values
type Gender uint64

const (
	Male   = 0
	Female = 1
)
