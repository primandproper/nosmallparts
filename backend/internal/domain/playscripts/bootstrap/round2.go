package bootstrap

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
)

// Round2ScriptCreation holds the input, cast members, and lines for creating a single play script.
type Round2ScriptCreation struct {
	Input       *playscripts.PlayScriptDatabaseCreationInput
	CastMembers []*playscripts.PlayScriptCastMemberDatabaseCreationInput
	Lines       []*playscripts.PlayScriptLineDatabaseCreationInput
}

// Round2Scripts returns all Round 2 play script creations (Shrek, Toy Story, Sandlot, Finding Nemo, Full House).
// Pass userID and accountID to populate CreatedByUser and BelongsToAccount on each input.
// Each element can be passed to CreatePlayScript as (c.Input, c.CastMembers, c.Lines).
func Round2Scripts(userID, accountID string) []Round2ScriptCreation {
	return []Round2ScriptCreation{
		{Input: round2ShrekInput(userID, accountID), CastMembers: round2ShrekCastMembers, Lines: round2ShrekLines},
		{Input: round2ToyStoryInput(userID, accountID), CastMembers: round2ToyStoryCastMembers, Lines: round2ToyStoryLines},
		{Input: round2SandlotInput(userID, accountID), CastMembers: round2SandlotCastMembers, Lines: round2SandlotLines},
		{Input: round2FindingNemoInput(userID, accountID), CastMembers: round2FindingNemoCastMembers, Lines: round2FindingNemoLines},
		{Input: round2FullHouseInput(userID, accountID), CastMembers: round2FullHouseCastMembers, Lines: round2FullHouseLines},
	}
}

// Round2Inputs returns a slice of every PlayScriptDatabaseCreationInput for Round 2 scripts.
func Round2Inputs(userID, accountID string) []*playscripts.PlayScriptDatabaseCreationInput {
	scripts := Round2Scripts(userID, accountID)
	result := make([]*playscripts.PlayScriptDatabaseCreationInput, len(scripts))
	for i := range scripts {
		result[i] = scripts[i].Input
	}
	return result
}

// --- Shrek ---

const (
	round2ShrekScriptID = "shrek_ogres_layers_01"
	round2ShrekDonkeyID = "shrek_donkey"
	round2ShrekShrekID  = "shrek_shrek"
)

var round2ShrekInputBase = &playscripts.PlayScriptDatabaseCreationInput{
	ID:               round2ShrekScriptID,
	Title:            "Shrek (Excerpt) - Ogres Have Layers",
	Description:      "Shrek and Donkey discuss ogres and onions. From Shrek (2001), screenplay by Ted Elliott & Terry Rossio.",
	Author:           "Ted Elliott, Terry Rossio",
	IsPublic:         true,
	CreatedByUser:    "REPLACE_USER_ID",
	BelongsToAccount: "REPLACE_ACCOUNT_ID",
}

var round2ShrekCastMembers = []*playscripts.PlayScriptCastMemberDatabaseCreationInput{
	{ID: round2ShrekDonkeyID, PlayScriptID: round2ShrekScriptID, Name: "Donkey", SortOrder: 0},
	{ID: round2ShrekShrekID, PlayScriptID: round2ShrekScriptID, Name: "Shrek", SortOrder: 1},
}

var round2ShrekLines = []*playscripts.PlayScriptLineDatabaseCreationInput{
	{ID: "shrek_01", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "Let me get this straight. You're gonna go fight a dragon and rescue a princess just so Farquaad will give you back a swamp which you only don't have because he filled it full of freaks in the first place. Is that about right?", LineIndex: 0, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_02", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekShrekID, LineText: "You know, maybe there's a good reason donkeys shouldn't talk.", LineIndex: 1, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_03", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "I don't get it. Why don't you just pull some of that ogre stuff on him? Throttle him, lay siege to his fortress, grinds his bones to make your bread, the whole ogre trip.", LineIndex: 2, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_04", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekShrekID, LineText: "Oh, I know what. Maybe I could have decapitated an entire village and put their heads on a pike, gotten a knife, cut open their spleen and drink their fluids. Does that sound good to you?", LineIndex: 3, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_05", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "Uh, no, not really, no.", LineIndex: 4, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_06", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekShrekID, LineText: "For your information, there's a lot more to ogres than people think.", LineIndex: 5, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_07", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "Example?", LineIndex: 6, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_08", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekShrekID, LineText: "Example? Okay, um, ogres are like onions.", Direction: "he holds out his onion", LineIndex: 7, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_09", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "They stink?", Direction: "sniffs the onion", LineIndex: 8, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_10", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekShrekID, LineText: "Yes - - No!", LineIndex: 9, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_11", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "They make you cry?", LineIndex: 10, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_12", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekShrekID, LineText: "No!", LineIndex: 11, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_13", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "You leave them in the sun, they get all brown, start sproutin' little white hairs.", LineIndex: 12, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_14", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekShrekID, LineText: "No! Layers! Onions have layers. Ogres have layers! Onions have layers. You get it? We both have layers.", Direction: "he heaves a sigh and then walks off", LineIndex: 13, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_15", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "Oh, you both have layers. Oh. You know, not everybody likes onions. Cake! Everybody loves cakes! Cakes have layers.", Direction: "trailing after Shrek", LineIndex: 14, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_16", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekShrekID, LineText: "I don't care... what everyone likes. Ogres are not like cakes.", LineIndex: 15, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_17", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "You know what else everybody likes? Parfaits. Have you ever met a person, you say, \"Let's get some parfait,\" they say, \"Hell no, I don't like no parfait\"? Parfaits are delicious.", LineIndex: 16, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_18", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekShrekID, LineText: "No! You dense, irritating, miniature beast of burden! Ogres are like onions! End of story. Bye-bye. See ya later.", LineIndex: 17, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_19", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "Parfaits may be the most delicious thing on the whole damn planet.", LineIndex: 18, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_20", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekShrekID, LineText: "You know, I think I preferred your humming.", LineIndex: 19, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "shrek_21", PlayScriptID: round2ShrekScriptID, CastMemberID: round2ShrekDonkeyID, LineText: "Do you have a tissue or something? I'm making a mess. Just the word parfait make me start slobbering.", LineIndex: 20, ScrollMode: playscripts.ScrollModeAuto},
}

func round2ShrekInput(userID, accountID string) *playscripts.PlayScriptDatabaseCreationInput {
	input := *round2ShrekInputBase
	input.CreatedByUser = userID
	input.BelongsToAccount = accountID
	return &input
}

// --- Toy Story ---

const (
	round2ToyStoryScriptID   = "toystory_youre_a_toy_01"
	round2ToyStoryWoodyID    = "toystory_woody"
	round2ToyStoryBuzzID     = "toystory_buzz"
	round2ToyStoryMrsDavisID = "toystory_mrs_davis"
	round2ToyStoryAndyID     = "toystory_andy"
)

var round2ToyStoryInputBase = &playscripts.PlayScriptDatabaseCreationInput{
	ID:               round2ToyStoryScriptID,
	Title:            "Toy Story (Excerpt) - You're a Toy!",
	Description:      "Woody and Buzz argue at the gas station. From Toy Story (1995), screenplay by Joss Whedon, Andrew Stanton, Joel Cohen, and Alec Sokolow.",
	Author:           "Joss Whedon, Andrew Stanton, Joel Cohen, Alec Sokolow",
	IsPublic:         true,
	CreatedByUser:    "REPLACE_USER_ID",
	BelongsToAccount: "REPLACE_ACCOUNT_ID",
}

var round2ToyStoryCastMembers = []*playscripts.PlayScriptCastMemberDatabaseCreationInput{
	{ID: round2ToyStoryWoodyID, PlayScriptID: round2ToyStoryScriptID, Name: "Woody", SortOrder: 0},
	{ID: round2ToyStoryBuzzID, PlayScriptID: round2ToyStoryScriptID, Name: "Buzz", SortOrder: 1},
	{ID: round2ToyStoryMrsDavisID, PlayScriptID: round2ToyStoryScriptID, Name: "Mrs. Davis", SortOrder: 2},
	{ID: round2ToyStoryAndyID, PlayScriptID: round2ToyStoryScriptID, Name: "Andy", SortOrder: 3},
}

var round2ToyStoryLines = []*playscripts.PlayScriptLineDatabaseCreationInput{
	{ID: "toystory_01", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "Buzz! You're alive! This is great! Oh, I'm saved! I'm saved! Andy'll find you here, he'll take us back to the room, and then you can tell everyone that this was all just a big mistake. Huh?", LineIndex: 0, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_02", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "Right? Buddy?", Direction: "weakly", LineIndex: 1, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_03", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryBuzzID, LineText: "I just want you to know that even though you tried to terminate me, revenge is not an idea we promote on my planet.", LineIndex: 2, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_04", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "Oh. Oh, that's good.", LineIndex: 3, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_05", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryBuzzID, LineText: "But we're not on my planet, are we?", Direction: "getting in Woody's face", LineIndex: 4, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_06", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "No.", LineIndex: 5, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_07", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "Ok! Come on! You want a piece of me?!", LineIndex: 6, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_08", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "Owwwww!!!", LineIndex: 7, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_09", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryMrsDavisID, LineText: "Next stop...", LineIndex: 8, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_10", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryAndyID, LineText: "Pizza Planet! Yeah!!!", LineIndex: 9, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_11", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "Andy?!", Direction: "GASP", LineIndex: 10, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_12", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "Doesn't he realize that I'm not there?", LineIndex: 11, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_13", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "I'm LOST! Oh, I'm a lost toy!", Direction: "beat", LineIndex: 12, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_14", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryBuzzID, LineText: "Buzz Lightyear Mission Log. The local Sheriff and I seem to be at a huge refueling station of some sort --", LineIndex: 13, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_15", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "You!", LineIndex: 14, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_16", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryBuzzID, LineText: "According to my nava-computer, the --", Direction: "into wrist communicator", LineIndex: 15, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_17", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "Shut up! Just shut up, you idiot!!", Direction: "whispering", LineIndex: 16, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_18", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryBuzzID, LineText: "Sheriff, this is no time to panic.", LineIndex: 17, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_19", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "This is the perfect time to panic! I'm lost, Andy is gone, they're going to move from their house in two days and it's all your fault!!", LineIndex: 18, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_20", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryBuzzID, LineText: "My fault?! If you hadn't pushed me out of the window in the first place --", LineIndex: 19, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_21", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "Oh, yeah? Well, if YOU hadn't shown up in your stupid little cardboard spaceship and taken away everything that was important to me --", LineIndex: 20, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_22", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryBuzzID, LineText: "Don't talk to me about importance. Because of YOU the security of this entire universe is in jeopardy.", LineIndex: 21, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_23", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "WHAT?!! What are you talking about?!", Direction: "incredulous", LineIndex: 22, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_24", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryBuzzID, LineText: "Right now, poised at the edge of the galaxy, Emperor Zurg has been secretly building a weapon with the destructive capacity to annihilate an entire planet. I alone have information that reveals this weapon's only weakness. And you, my friend, are responsible for delaying my rendez-vous with Star Command.", Direction: "pointing at Woody", LineIndex: 23, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_25", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "YOU ARE A TOY!!! You aren't the real Buzz Lightyear, you're an action figure!! You are a child's plaything!!!", LineIndex: 24, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_26", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryBuzzID, LineText: "You are a sad strange little man and you have my pity. Farewell.", LineIndex: 25, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "toystory_27", PlayScriptID: round2ToyStoryScriptID, CastMemberID: round2ToyStoryWoodyID, LineText: "Oh, yeah? Well, good riddance, ya looney!", LineIndex: 26, ScrollMode: playscripts.ScrollModeAuto},
}

func round2ToyStoryInput(userID, accountID string) *playscripts.PlayScriptDatabaseCreationInput {
	input := *round2ToyStoryInputBase
	input.CreatedByUser = userID
	input.BelongsToAccount = accountID
	return &input
}

// --- Sandlot ---

const (
	round2SandlotScriptID   = "sandlot_killin_me_smalls_01"
	round2SandlotBennyID    = "sandlot_benny"
	round2SandlotBertramID  = "sandlot_bertram"
	round2SandlotScottyID   = "sandlot_scotty"
	round2SandlotHamID      = "sandlot_ham"
	round2SandlotDenunezID  = "sandlot_denunez"
	round2SandlotYeahYeahID = "sandlot_yeah_yeah"
	round2SandlotTimmyID    = "sandlot_timmy"
	round2SandlotGangID     = "sandlot_gang"
	round2SandlotSquintsID  = "sandlot_squints"
)

var round2SandlotInputBase = &playscripts.PlayScriptDatabaseCreationInput{
	ID:               round2SandlotScriptID,
	Title:            "The Sandlot (Excerpt) - You're Killin' Me Smalls",
	Description:      "The gang introduces Scotty to chewing tobacco. From The Sandlot (1993), screenplay by David Mickey Evans and Robert Gunter.",
	Author:           "David Mickey Evans, Robert Gunter",
	IsPublic:         true,
	CreatedByUser:    "REPLACE_USER_ID",
	BelongsToAccount: "REPLACE_ACCOUNT_ID",
}

var round2SandlotCastMembers = []*playscripts.PlayScriptCastMemberDatabaseCreationInput{
	{ID: round2SandlotBennyID, PlayScriptID: round2SandlotScriptID, Name: "Benny", SortOrder: 0},
	{ID: round2SandlotBertramID, PlayScriptID: round2SandlotScriptID, Name: "Bertram", SortOrder: 1},
	{ID: round2SandlotScottyID, PlayScriptID: round2SandlotScriptID, Name: "Scotty", SortOrder: 2},
	{ID: round2SandlotHamID, PlayScriptID: round2SandlotScriptID, Name: "Ham", SortOrder: 3},
	{ID: round2SandlotDenunezID, PlayScriptID: round2SandlotScriptID, Name: "Denunez", SortOrder: 4},
	{ID: round2SandlotYeahYeahID, PlayScriptID: round2SandlotScriptID, Name: "Yeah-Yeah", SortOrder: 5},
	{ID: round2SandlotTimmyID, PlayScriptID: round2SandlotScriptID, Name: "Timmy", SortOrder: 6},
	{ID: round2SandlotGangID, PlayScriptID: round2SandlotScriptID, Name: "Gang", SortOrder: 7},
	{ID: round2SandlotSquintsID, PlayScriptID: round2SandlotScriptID, Name: "Squints", SortOrder: 8},
}

var round2SandlotLines = []*playscripts.PlayScriptLineDatabaseCreationInput{
	{ID: "sandlot_01", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotBennyID, LineText: "Crack! Boom! Outta here! You see the looks on their faces? Did ya? It was like, \"Duh... so that's how you play baseball.\"", LineIndex: 0, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_02", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotBertramID, LineText: "Crap! I almost forgot. Chaw!", LineIndex: 1, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_03", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotBertramID, LineText: "I was savin' it for a good time.", LineIndex: 2, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_04", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotScottyID, LineText: "What is it?", LineIndex: 3, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_05", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotHamID, LineText: "Geez, Smalls... I s'pose you don't who The Babe is either. It's chaw. Plug. Wad. Chewing Tobacco!", LineIndex: 4, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_06", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotScottyID, LineText: "Whadda you do with it?", LineIndex: 5, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_07", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotHamID, LineText: "You're killin' me, Smalls. You chew it, of course.", LineIndex: 6, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_08", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotDenunezID, LineText: "All the pros do.", LineIndex: 7, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_09", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotYeahYeahID, LineText: "Yeah-yeah, gives ya tons of energy.", LineIndex: 8, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_10", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotTimmyID, LineText: "Let's dip.", LineIndex: 9, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_11", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotGangID, LineText: "Let's dip.", LineIndex: 10, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_12", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotBennyID, LineText: "Smooth.", LineIndex: 11, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_13", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotHamID, LineText: "Juicy.", LineIndex: 12, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_14", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotSquintsID, LineText: "Tangy.", LineIndex: 13, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_15", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotScottyID, LineText: "Kinda tastes like an ashtray smells.", LineIndex: 14, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_16", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotDenunezID, LineText: "Supposed to.", LineIndex: 15, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "sandlot_17", PlayScriptID: round2SandlotScriptID, CastMemberID: round2SandlotSquintsID, LineText: "Let's ride.", LineIndex: 16, ScrollMode: playscripts.ScrollModeAuto},
}

func round2SandlotInput(userID, accountID string) *playscripts.PlayScriptDatabaseCreationInput {
	input := *round2SandlotInputBase
	input.CreatedByUser = userID
	input.BelongsToAccount = accountID
	return &input
}

// --- Finding Nemo ---

const (
	round2FindingNemoScriptID = "findingnemo_just_keep_swimming_01"
	round2FindingNemoMarlinID = "findingnemo_marlin"
	round2FindingNemoDoryID   = "findingnemo_dory"
	round2FindingNemoBothID   = "findingnemo_both"
)

var round2FindingNemoInputBase = &playscripts.PlayScriptDatabaseCreationInput{
	ID:               round2FindingNemoScriptID,
	Title:            "Finding Nemo (Excerpt) - Just Keep Swimming",
	Description:      "Marlin and Dory in the whale and reading the mask. From Finding Nemo (2003), screenplay by Andrew Stanton, Bob Peterson, and David Reynolds.",
	Author:           "Andrew Stanton, Bob Peterson, David Reynolds",
	IsPublic:         true,
	CreatedByUser:    "REPLACE_USER_ID",
	BelongsToAccount: "REPLACE_ACCOUNT_ID",
}

var round2FindingNemoCastMembers = []*playscripts.PlayScriptCastMemberDatabaseCreationInput{
	{ID: round2FindingNemoMarlinID, PlayScriptID: round2FindingNemoScriptID, Name: "Marlin", SortOrder: 0},
	{ID: round2FindingNemoDoryID, PlayScriptID: round2FindingNemoScriptID, Name: "Dory", SortOrder: 1},
	{ID: round2FindingNemoBothID, PlayScriptID: round2FindingNemoScriptID, Name: "Both", SortOrder: 2},
}

var round2FindingNemoLines = []*playscripts.PlayScriptLineDatabaseCreationInput{
	{ID: "findingnemo_01", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Nemo. Nemo.", Direction: "mutters", LineIndex: 0, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_02", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Are you gonna eat that? Careful with that hammer...", LineIndex: 1, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_03", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Huh? No, no! What does it say? Dory!", LineIndex: 2, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_04", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Sea monkey has my money...", LineIndex: 3, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_05", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Wake up! Get up! Come on! Come on!", LineIndex: 4, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_06", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Yes, I'm a natural blue...", LineIndex: 5, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_07", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Get up!", LineIndex: 6, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_08", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Look out! Sharks eat fish! Aaaaaah!", LineIndex: 7, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_09", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoBothID, LineText: "AAAAAAAAAAHHH!!!", LineIndex: 8, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_10", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Wow. Dusty.", LineIndex: 9, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_11", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "The mask! Where's the mask? No! No, not the mask! Get it! Get the mask! Get it!", Direction: "gasps", LineIndex: 10, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_12", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Hoo doot doo doot doot doo doot. Whoo-hoo! La la la la la la. Just keeps going on, doesn't it? Echo! Echo! Hey, what are you doing?", Direction: "singing", LineIndex: 11, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_13", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "It's gone. I've lost the mask.", LineIndex: 12, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_14", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Did you drop it?", LineIndex: 13, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_15", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "You dropped it! That was my only chance of finding my son, now it's gone.", LineIndex: 14, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_16", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Hey, Mr. Grumpy Gills. When life gets you down, you know what you gotta do?", LineIndex: 15, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_17", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "I don't wanna know what you gotta do when life gets you down.", LineIndex: 16, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_18", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Just keep swimming. Just keep swimming, swimming, swimming. What do we do? We swim, swim.", Direction: "singing", LineIndex: 17, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_19", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Dory, no singing.", LineIndex: 18, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_20", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Ho ho ho ho ho ho! I love to swim! When you want to swim..", Direction: "singing", LineIndex: 19, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_21", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "See, I'm going to get stuck now with that song now it's in my head!", LineIndex: 20, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_22", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Sorry.", LineIndex: 21, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_23", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Dory, do you see anything?", LineIndex: 22, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_24", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Aaah! Something's got me!", LineIndex: 23, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_25", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "That was me. I'm sorry.", LineIndex: 24, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_26", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Who was that?", Direction: "gasps", LineIndex: 25, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_27", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Who could it be? It's me!", LineIndex: 26, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_28", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Are..are you my conscience?", LineIndex: 27, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_29", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Yeah, yeah. I'm your conscience. We haven't spoken for a while. How are you?", LineIndex: 28, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_30", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Hmm, can't complain.", LineIndex: 29, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_31", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Yeah? Good. Now, Dory. I want you to tell me..do you see anything?", LineIndex: 30, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_32", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "I see..I see a light.", LineIndex: 31, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_33", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "A light.", LineIndex: 32, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_34", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Yeah. Over there. Hey, conscience. Am I dead?", LineIndex: 33, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_35", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "No, I see it too. What is it?", LineIndex: 34, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_36", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "It's so pretty.", LineIndex: 35, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_37", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "I'm feeling...happy. Which is a big deal for me.", LineIndex: 36, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_38", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "I want to touch it. Oh!", LineIndex: 37, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_39", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Hey, come back. Come on back here.", LineIndex: 38, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_40", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "I'm gonna get you. I'm gonna get you. I'm gonna swim with you.", Direction: "singing", LineIndex: 39, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_41", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "I'm gonna get you. I'm gonna be your best friend...good feeling's gone.", LineIndex: 40, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_42", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "I can't see! I don't know where I'm going!", LineIndex: 41, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_43", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Haah!", LineIndex: 42, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_44", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "The mask!", LineIndex: 43, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_45", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "What mask?", LineIndex: 44, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_46", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Okay, I can't see a thing.", LineIndex: 45, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_47", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Oh, gee!", LineIndex: 46, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_48", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Hey, look! A mask!", LineIndex: 47, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_49", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Read it!", LineIndex: 48, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_50", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "I'm sorry, but if you could just bring it a little closer, I kind of need the light. That's great, keep it right there.", LineIndex: 49, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_51", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Just read it!", LineIndex: 50, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_52", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Okay, okay. Mr. Bossy. Uh, 'P'. Okay, 'P'. 'Shh-eer...Sher--P. Sher--P. Shirley? P.--'. Oh! The first line's 'P. Sherman'!", LineIndex: 51, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_53", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "P. Sherman doesn't make any sense!", LineIndex: 52, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_54", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Okay, second line. '42'.", LineIndex: 53, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_55", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Don't eat me! Don't eat me! Aaaah!", LineIndex: 54, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_56", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Light, please. 'Walla--Walla--Walla-beee'...", LineIndex: 55, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_57", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Waah! Waaah! Waaaah!", LineIndex: 56, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_58", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "The second line's '42 Wallaby Way'!", LineIndex: 57, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_59", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "That's great! Speed read! Take a guess! No pressure! No problem! There's a lot of pressure! Pressure! Take a guess now with pressure!", LineIndex: 58, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_60", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "'Sydney'. It's 'Sydney'!", LineIndex: 59, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_61", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Duck!", LineIndex: 60, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_62", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "Aaah!", LineIndex: 61, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_63", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "I'm dead, I'm dead, I'm dead, I'm dead, I'm dead, I died, I'm dead.", LineIndex: 62, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_64", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Whoo-hoo! We did it, we did it! Oh yeah, yeah, yeah! No eating here tonight, whoo!", Direction: "singing", LineIndex: 63, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_65", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoBothID, LineText: "Eating here tonight!", Direction: "singing", LineIndex: 64, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_66", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Dory.", LineIndex: 65, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_67", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "No, no, no eating here tonight. You on a diet--", Direction: "singing", LineIndex: 66, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_68", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Dory! What did the mask say?", LineIndex: 67, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_69", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "'P. Sherman, 42 Wallaby Way, Sydney'. I remember what it said! I usually forget things, but I remembered it this time!", Direction: "gasps", LineIndex: 68, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_70", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoMarlinID, LineText: "Whoa, whoa, wait! Where is that?", LineIndex: 69, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_71", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "I don't know. But who cares? I remembered!", LineIndex: 70, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_72", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoBothID, LineText: "Aaah!", LineIndex: 71, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "findingnemo_73", PlayScriptID: round2FindingNemoScriptID, CastMemberID: round2FindingNemoDoryID, LineText: "P. Sherman, 42 Wallaby Way, Sydney. I remembered it again!", LineIndex: 72, ScrollMode: playscripts.ScrollModeAuto},
}

func round2FindingNemoInput(userID, accountID string) *playscripts.PlayScriptDatabaseCreationInput {
	input := *round2FindingNemoInputBase
	input.CreatedByUser = userID
	input.BelongsToAccount = accountID
	return &input
}

// --- Full House ---

const (
	round2FullHouseScriptID = "fullhouse_how_rude_01"
	round2FullHouseJoeyID   = "fullhouse_joey"
	round2FullHouseDJID     = "fullhouse_dj"
	round2FullHouseJesseID  = "fullhouse_jesse"
)

var round2FullHouseInputBase = &playscripts.PlayScriptDatabaseCreationInput{
	ID:               round2FullHouseScriptID,
	Title:            "Full House (Excerpt) - Sisters & Commercial",
	Description:      "Joey, Jesse and DJ talk about Stephanie stealing DJ's commercial. From Full House (1987-1995), created by Jeff Franklin.",
	Author:           "Jeff Franklin",
	IsPublic:         true,
	CreatedByUser:    "REPLACE_USER_ID",
	BelongsToAccount: "REPLACE_ACCOUNT_ID",
}

var round2FullHouseCastMembers = []*playscripts.PlayScriptCastMemberDatabaseCreationInput{
	{ID: round2FullHouseJoeyID, PlayScriptID: round2FullHouseScriptID, Name: "Joey", SortOrder: 0},
	{ID: round2FullHouseDJID, PlayScriptID: round2FullHouseScriptID, Name: "DJ", SortOrder: 1},
	{ID: round2FullHouseJesseID, PlayScriptID: round2FullHouseScriptID, Name: "Jesse", SortOrder: 2},
}

var round2FullHouseLines = []*playscripts.PlayScriptLineDatabaseCreationInput{
	{ID: "fullhouse_01", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "DJ, we need to talk.", LineIndex: 0, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_02", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseDJID, LineText: "Let me save you a lot of trouble. You're right, I'm wrong, and I don't care.", LineIndex: 1, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_03", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "Listen, DJ Jesse and I strike out on auditions all the time. Being a performer is basically a series of rejections occasionally interrupted by work.", LineIndex: 2, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_04", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseDJID, LineText: "Oh, yeah? Well, did either of you have a part stolen by your little sister?", LineIndex: 3, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_05", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJesseID, LineText: "You know something? She's r--DJ, you're right. Stephanie stole that part right out from under you. In fact, she planned this whole thing. She tricked us into bringing her along, pretending that she cares about you.", LineIndex: 4, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_06", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "The conniving little sneak.", LineIndex: 5, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_07", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJesseID, LineText: "And then you fell into Stephanie's fiendish little trap by doing a great audition.", LineIndex: 6, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_08", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "The monster.", LineIndex: 7, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_09", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJesseID, LineText: "Meanwhile, the sinister little Stephanie...used her telekinetic powers to send us out of the room so she could steal your commercial.", LineIndex: 8, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_10", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "I say we light some torches, find the she-devil and drive her into the night!", LineIndex: 9, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_11", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseDJID, LineText: "Okay, you guys were right. She didn't do it on purpose. But I'm still mad at her.", LineIndex: 10, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_12", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "All right. All right, kiddo. I wanna know what's eating you.", LineIndex: 11, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_13", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJesseID, LineText: "All right, pal. Come on, what's the matter? Tell us.", LineIndex: 12, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_14", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseDJID, LineText: "I don't know. I'm so sick of sisters. Everything was just fine when it was just me. And then Stephanie came along, and everyone said she was so cute. Now everyone says Michelle is so cute. It's so disgusting.", LineIndex: 13, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_15", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "Was it so disgusting when you did the play and everyone thought you were so cute?", LineIndex: 14, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_16", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseDJID, LineText: "No, that was fine. Until one of the so-cute sisters ruined everything again.", LineIndex: 15, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_17", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJesseID, LineText: "Buddy. It's tough being the oldest. But you can't overlook the benefits. I mean, you get everything new. No hand-me-downs.", LineIndex: 16, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_18", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "And you get to do everything first.", LineIndex: 17, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_19", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJesseID, LineText: "That's right. Date.", LineIndex: 18, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_20", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "Drive.", LineIndex: 19, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_21", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJesseID, LineText: "You're the first to be admitted to an R-rated picture.", LineIndex: 20, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_22", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "Without a parent or legal guardian.", LineIndex: 21, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_23", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJesseID, LineText: "But the best part is...you got two little sisters who need you...and look up to you...and love you very, very much.", LineIndex: 22, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_24", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "So, what do you think? Should we light the torches...find the beast and drive her into the darkness?", LineIndex: 23, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_25", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseDJID, LineText: "I guess not. But if we did, I'd still have one little sister left.", LineIndex: 24, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_26", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJesseID, LineText: "Noogie!", LineIndex: 25, ScrollMode: playscripts.ScrollModeAuto},
	{ID: "fullhouse_27", PlayScriptID: round2FullHouseScriptID, CastMemberID: round2FullHouseJoeyID, LineText: "Double noogie!", LineIndex: 26, ScrollMode: playscripts.ScrollModeAuto},
}

func round2FullHouseInput(userID, accountID string) *playscripts.PlayScriptDatabaseCreationInput {
	input := *round2FullHouseInputBase
	input.CreatedByUser = userID
	input.BelongsToAccount = accountID
	return &input
}
