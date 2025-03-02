package magic

import (
	"fmt"
	"unsafe"
)

const (
	NBKeys   = 1 << 22   // max, adjust to arbitrate between speed (collisions) & memory, SHOULD BE A POWER OF TWO !
	NBValues = 256 * 256 // max, adjustable, less than nbkeys. Not necessarily a power of two.
)

// There is ONE big table that maps the sqt (combination of squere and table) and the occupancy data,
// to a uint8, the index of the value in the dictionnary.
// No limit on number of keys.
// keys should NEVER be 2^64-1 !
// No more than 256 distinct VALUES per sqt (ie, a given square and table), forming a chunk.
// Dictionnary is made of up to 256 chunks, 1 per sqt.
// Chunks may overlapp if less than 256.
// NB : fields are public to facilitate save/load, but should never be acessed directly...
type MagicMap struct {
	// Key search
	Key    [NBKeys]uint64 // fixed array of keys, addressed with a hash and linear search
	Sqt    [NBKeys]uint8  // fixed array, in sync with keys, combined square and what to detect collisions
	Vindex [NBKeys]uint8  // fixed array, in sync with keys, pointing to dictionnary value index.

	// value dictionnay - split in up to 256 chunks of up to 256 distinct Values for a given sqt
	Values [NBValues]uint64 // value dictionary, fixed array of all output values.
	// chunks always have a length of 256, but may overlapp, in particular for empty chunks.
	Start [256]uint64 // points to chunk start in the values array, indexed on sqt - not necessarily a power of two.  Not neccesarily in order.
}

// magic numbers used by hash function, based on reduced 4-bit square
var magicNumbers = [...]uint64{ // 256 prime numbers to choose from based upon sqt value
	45317, 45319, 45329, 45337, 45341, 45343, 45361, 45377, 45389, 45403,
	45413, 45427, 45433, 45439, 45481, 45491, 45497, 45503, 45523, 45533,
	45541, 45553, 45557, 45569, 45587, 45589, 45599, 45613, 45631, 45641,
	45659, 45667, 45673, 45677, 45691, 45697, 45707, 45737, 45751, 45757,
	45763, 45767, 45779, 45817, 45821, 45823, 45827, 45833, 45841, 45853,
	45863, 45869, 45887, 45893, 45943, 45949, 45953, 45959, 45971, 45979,
	45989, 46021, 46027, 46049, 46051, 46061, 46073, 46091, 46093, 46099,
	46103, 46133, 46141, 46147, 46153, 46171, 46181, 46183, 46187, 46199,
	46997, 47017, 47041, 47051, 47057, 47059, 47087, 47093, 47111, 47119,
	47123, 47129, 47137, 47143, 47147, 47149, 47161, 47189, 47207, 47221,

	47237, 47251, 47269, 47279, 47287, 47293, 47297, 47303, 47309, 47317,
	47339, 47351, 47353, 47363, 47381, 47387, 47389, 47407, 47417, 47419,
	47431, 47441, 47459, 47491, 47497, 47501, 47507, 47513, 47521, 47527,
	47533, 47543, 47563, 47569, 47581, 47591, 47599, 47609, 47623, 47629,
	47639, 47653, 47657, 47659, 47681, 47699, 47701, 47711, 47713, 47717,
	47737, 47741, 47743, 47777, 47779, 47791, 47797, 47807, 47809, 47819,
	47837, 47843, 47857, 47869, 47881, 47903, 47911, 47917, 47933, 47939,
	47947, 47951, 47963, 47969, 47977, 47981, 48017, 48023, 48029, 48049,
	48311, 48313, 48337, 48341, 48353, 48371, 48383, 48397, 48407, 48409,
	48413, 48437, 48449, 48463, 48473, 48479, 48481, 48487, 48491, 48497,

	48523, 48527, 48533, 48539, 48541, 48563, 48571, 48589, 48593, 48611,
	48619, 48623, 48647, 48649, 48661, 48673, 48677, 48679, 48731, 48733,
	48751, 48757, 48761, 48767, 48779, 48781, 48787, 48799, 48809, 48817,
	48821, 48823, 48847, 48857, 48859, 48869, 48871, 48883, 48889, 48907,
	48947, 48953, 48973, 48989, 48991, 49003, 49009, 49019, 49031, 49033,

	49037, 49043, 49057, 49069, 49081, 49103,

	/* some other prime numbers if needed ...
	49139,  49157,  49169,  49171,  49177,  49193,  49199,  49201,  49207,  49211,
	49223,  49253,  49261,  49277,  49279,  49297,  49307,  49331,  49333,  49339,
	49363,  49367,  49369,  49391,  49393,  49409,  49411,  49417,  49429,  49433,
	49451,  49459,  49463,  49477,  49481,  49499,  49523,  49529,  49531,  49537,
	49547,  49549,  49559,  49597,  49603,  49613,  49627,  49633,  49639,  49663,
	49667,  49669,  49681,  49697,  49711,  49727,  49739,  49741,  49747,  49757,
	49783,  49787,  49789,  49801,  49807,  49811,  49823,  49831,  49843,  49853,
	49871,  49877,  49891,  49919,  49921,  49927,  49937,  49939,  49943,  49957,
	49991,  49993,  49999,  50021,  50023,  50033,  50047,  50051,  50053,  50069,
	50077,  50087,  50093,  50101,  50111,  50119,  50123,  50129,  50131,  50147,
	50153,  50159,  50177,  50207,  50221,  50227,  50231,  50261,  50263,  50273,
	50287,  50291,  50311,  50321,  50329,  50333,  50341,  50359,  50363,  50377,
	50383,  50387,  50411,  50417,  50423,  50441,  50459,  50461,  50497,  50503,
	50513,  50527,  50539,  50543,  50549,  50551,  50581,  50587,  50591,  50593,
	50599,  50627,  50647,  50651,  50671,  50683,  50707,  50723,  50741,  50753,
	50767,  50773,  50777,  50789,  50821,  50833,  50839,  50849,  50857,  50867,
	50873,  50891,  50893,  50909,  50923,  50929,  50951,  50957,  50969,  50971,
	50989,  50993,  51001,  51031,  51043,  51047,  51059,  51061,  51071,  51109,
	51131,  51133,  51137,  51151,  51157,  51169,  51193,  51197,  51199,  51203,
	51217,  51229,  51239,  51241,  51257,  51263,  51283,  51287,  51307,  51329,
	51341,  51343,  51347,  51349,  51361,  51383,  51407,  51413,  51419,  51421,
	51427,  51431,  51437,  51439,  51449,  51461,  51473,  51479,  51481,  51487,
	51503,  51511,  51517,  51521,  51539,  51551,  51563,  51577,  51581,  51593,
	51599,  51607,  51613,  51631,  51637,  51647,  51659,  51673,  51679,  51683,
	51691,  51713,  51719,  51721,  51749,  51767,  51769,  51787,  51797,  51803,
	51817,  51827,  51829,  51839,  51853,  51859,  51869,  51871,  51893,  51899,
	51907,  51913,  51929,  51941,  51949,  51971,  51973,  51977,  51991,  52009,
	52021,  52027,  52051,  52057,  52067,  52069,  52081,  52103,  52121,  52127,
	52147,  52153,  52163,  52177,  52181,  52183,  52189,  52201,  52223,  52237,
	52249,  52253,  52259,  52267,  52289,  52291,  52301,  52313,  52321,  52361,
	52363,  52369,  52379,  52387,  52391,  52433,  52453,  52457,  52489,  52501,
	52511,  52517,  52529,  52541,  52543,  52553,  52561,  52567,  52571,  52579,
	52583,  52609,  52627,  52631,  52639,  52667,  52673,  52691,  52697,  52709,
	52711,  52721,  52727,  52733,  52747,  52757,  52769,  52783,  52807,  52813,
	52817,  52837,  52859,  52861,  52879,  52883,  52889,  52901,  52903,  52919,
	52937,  52951,  52957,  52963,  52967,  52973,  52981,  52999,  53003,  53017,
	53047,  53051,  53069,  53077,  53087,  53089,  53093,  53101,  53113,  53117,
	53129,  53147,  53149,  53161,  53171,  53173,  53189,  53197,  53201,  53231,
	53233,  53239,  53267,  53269,  53279,  53281,  53299,  53309,  53323,  53327,
	53353,  53359,  53377,  53381,  53401,  53407,  53411,  53419,  53437,  53441,
	53453,  53479,  53503,  53507,  53527,  53549,  53551,  53569,  53591,  53593,
	53597,  53609,  53611,  53617,  53623,  53629,  53633,  53639,  53653,  53657,
	53681,  53693,  53699,  53717,  53719,  53731,  53759,  53773,  53777,  53783,
	53791,  53813,  53819,  53831,  53849,  53857,  53861,  53881,  53887,  53891,
	53897,  53899,  53917,  53923,  53927,  53939,  53951,  53959,  53987,  53993,
	54001,  54011,  54013,  54037,  54049,  54059,  54083,  54091,  54101,  54121,
	54133,  54139,  54151,  54163,  54167,  54181,  54193,  54217,  54251,  54269,
	54277,  54287,  54293,  54311,  54319,  54323,  54331,  54347,  54361,  54367,
	54371,  54377,  54401,  54403,  54409,  54413,  54419,  54421,  54437,  54443,
	54449,  54469,  54493,  54497,  54499,  54503,  54517,  54521,  54539,  54541,
	54547,  54559,  54563,  54577,  54581,  54583,  54601,  54617,  54623,  54629,
	54631,  54647,  54667,  54673,  54679,  54709,  54713,  54721,  54727,  54751,
	54767,  54773,  54779,  54787,  54799,  54829,  54833,  54851,  54869,  54877,
	54881,  54907,  54917,  54919,  54941,  54949,  54959,  54973,  54979,  54983,
	55001,  55009,  55021,  55049,  55051,  55057,  55061,  55073,  55079,  55103,
	55109,  55117,  55127,  55147,  55163,  55171,  55201,  55207,  55213,  55217,
	55219,  55229,  55243,  55249,  55259,  55291,  55313,  55331,  55333,  55337,
	55339,  55343,  55351,  55373,  55381,  55399,  55411,  55439,  55441,  55457,
	55469,  55487,  55501,  55511,  55529,  55541,  55547,  55579,  55589,  55603,
	55609,  55619,  55621,  55631,  55633,  55639,  55661,  55663,  55667,  55673,
	55681,  55691,  55697,  55711,  55717,  55721,  55733,  55763,  55787,  55793,
	55799,  55807,  55813,  55817,  55819,  55823,  55829,  55837,  55843,  55849,
	55871,  55889,  55897,  55901,  55903,  55921,  55927,  55931,  55933,  55949,
	55967,  55987,  55997,  56003,  56009,  56039,  56041,  56053,  56081,  56087,
	56093,  56099,  56101,  56113,  56123,  56131,  56149,  56167,  56171,  56179,
	56197,  56207,  56209,  56237,  56239,  56249,  56263,  56267,  56269,  56299,
	56311,  56333,  56359,  56369,  56377,  56383,  56393,  56401,  56417,  56431,
	56437,  56443,  56453,  56467,  56473,  56477,  56479,  56489,  56501,  56503,
	56509,  56519,  56527,  56531,  56533,  56543,  56569,  56591,  56597,  56599,
	56611,  56629,  56633,  56659,  56663,  56671,  56681,  56687,  56701,  56711,
	56713,  56731,  56737,  56747,  56767,  56773,  56779,  56783,  56807,  56809,
	56813,  56821,  56827,  56843,  56857,  56873,  56891,  56893,  56897,  56909,
	56911,  56921,  56923,  56929,  56941,  56951,  56957,  56963,  56983,  56989,
	56993,  56999,  57037,  57041,  57047,  57059,  57073,  57077,  57089,  57097,
	57107,  57119,  57131,  57139,  57143,  57149,  57163,  57173,  57179,  57191,
	57193,  57203,  57221,  57223,  57241,  57251,  57259,  57269,  57271,  57283,
	*/
}

// Key, in hash, should never be 0.
// sqt is a combined square and piece in 1 byte.
func hash(sqt uint8, key uint64) uint64 {
	return (magicNumbers[sqt] * (key))
}

// Compute the power of two equal or greater than v.
// By convention, 0 returns 0 and not 1.
func NextPowerOfTwo(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	v++
	return v
}

// ==============================================
// MagicMap usage
//===============================================

// Retrieve a uint64 value given a 64-bits key and a uint8 combination of Square and What.
// CAUTION : will loop forever if key does not exists in table !
func (m MagicMap) Get(sqt uint8, key uint64) uint64 {
	// inverse key to allow for 0-key !
	key = ^key
	// search for matching key - infinite loop while not found ...
	for i := hash(sqt, key) & (NBKeys - 1); ; i = (i + 1) & (NBKeys - 1) {
		if (m.Key[i] == key) && (m.Sqt[i] == sqt) { // found matching key and sqt !
			// return value
			return m.Values[uint64(m.Vindex[i])+m.Start[sqt]]
		}
	}
}

// ==============================
// MagicMap construction
// ==============================

type TableEntry struct {
	Sqt    uint8             // square and piece to be implemented by the table
	Values map[uint64]uint64 // maps occ keys to values for given sqt
}

type Stats struct {
	CollCount     uint64 // number of collisions in keys
	CollSumSearch uint64 // sum of all search distance on collision
	CollMaxSearch uint64 // maximum search distance when collision
	ActualKeys    uint64 // number of keys actually used
	ActualValues  uint64 // number of output values actually used
	MemoryUsed    uint64 // memory used by the table
}

func (st Stats) String() string {
	return fmt.Sprintf(
		`
  ------------------
  Stats for MagicMap 
  ------------------
  CollCount           %d
  CollSumSearch       %d
  CollAverageSearch   %.2f
  CollMaxSearch       %d
  ActualKeys          %d  / %d
  ActualValues        %d  / %d
  MemoryUsed          %d bytes
  -----------------------------------
  `,
		st.CollCount, st.CollSumSearch, float64(st.CollSumSearch)/float64(st.CollCount), st.CollMaxSearch,
		st.ActualKeys, NBKeys,
		st.ActualValues, NBValues,
		st.MemoryUsed)
}

// Init a  MagicMap table with various entries.
// Each TableEntry will create a new Chunk, in the order they are provided.
// If a chunk does not exist, it will start at 0. It should NEVER be called !
// Caution : keys are stored inverted, so no key should ever be 64 ones (^uint64(0)) !
// Caution : result is NOT deterministic, because we directly range over map entries.
func InitMagicMap(m *MagicMap, te ...TableEntry) (stat Stats) {

	if m == nil {
		panic("destination MagicMap is required to create")
	}

	// measure memory footprint
	stat.MemoryUsed = uint64(unsafe.Sizeof(*m))

	if NextPowerOfTwo(NBKeys) != NBKeys {
		panic(fmt.Sprintf("NBKeys (%d) is not a power of 2", NBKeys))
	}

	// check inputs
	if len(te) == 0 {
		return
	}
	if len(te) > 256 {
		panic(fmt.Sprintf("too many Sqt (%d tables - max is 256)", len(te)))
	}

	// create/init internal temporary data structures
	type dicentry struct {
		v   uint64
		sqt uint8
	}
	dictionnary := make(map[dicentry]uint8, NBValues) // value,sqt  -> relative index to start of chunk
	countinputkeys := 0

	// === Prepare dictionnary for output values
	for _, table := range te {

		// count input keys
		countinputkeys += len(table.Values)

		// update dictionnary chunk starts
		if m.Start[table.Sqt] != 0 {
			panic(fmt.Sprintf("Sqt 0x%X already exists - duplicated TableEntries ?", table.Sqt))
		}
		m.Start[table.Sqt] = uint64(len(dictionnary))

		// add all sqt/values pairs into dictionnary
		for _, v := range table.Values {
			if _, ok := dictionnary[dicentry{v: v, sqt: table.Sqt}]; !ok {
				relIdx := uint64(len(dictionnary)) - m.Start[table.Sqt]
				if relIdx >= 256 {
					panic(fmt.Sprintf("too many output values for sqt 0x%X", table.Sqt))
				}
				dictionnary[dicentry{v: v, sqt: table.Sqt}] = uint8(relIdx)
			}
		}
	}

	// check again ...
	if len(dictionnary) > NBValues {
		panic(fmt.Sprintf("too many total output values (%d) - configurated maximum was %d", len(dictionnary), NBValues))
	}
	if countinputkeys >= NBKeys { // because 1 key cannot be used ...
		panic(fmt.Sprintf("too many total input keys (%d) - allowed maximum was (%d-1)", countinputkeys, NBKeys))
	}

	// == fill in input keys
	for _, table := range te {
		for k, v := range table.Values {
			k = ^k // invert k !
			h := hash(table.Sqt, k) & (NBKeys - 1)
			var s uint64 // search distance
			// endless loop until empty slot is found
			for i := h; s < NBKeys; s, i = s+1, (i+1)&(NBKeys-1) {
				if m.Key[i] == 0 { // found available slot
					m.Key[i] = k
					m.Sqt[i] = table.Sqt
					m.Vindex[i] = dictionnary[dicentry{v: v, sqt: table.Sqt}]
					break
				}
			}
			if s != 0 {
				stat.CollCount++
				stat.CollSumSearch += s
				stat.CollMaxSearch = max(stat.CollMaxSearch, s)
			}
		}
	}

	// == fill in output values
	for de, ri := range dictionnary {
		m.Values[uint64(ri)+m.Start[de.sqt]] = de.v
	}

	// Update stats
	stat.ActualKeys = uint64(countinputkeys)
	stat.ActualValues = uint64(len(dictionnary))

	return stat
}
