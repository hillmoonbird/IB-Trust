package main

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

// CurrentBlockVersion is the version of the block metadata
const CurrentBlockVersion = 1

// GenesisBlockPreviousBlockHash is the hard-coded canonical stand-in hash of the non-existent previous block
const GenesisBlockPreviousBlockHash = "1000000000000000000000000000000000000000000000000000000000000001"

// ChainParams describe the genesis block and other blockchain properties
var defaultChainParams = ChainParams{
	ConsensusType:             ChainConsensusPoA,
	GenesisBlockHash:          "9a0ff19183d1525a36de803047de4b73eb72506be8c81296eb463476a5c2d9e2",
	GenesisBlockHashSignature: "30460221008b8b3b3cfee2493ef58f2f6a1f1768b564f4c9e9a341ad42912cbbcf5c3ec82f022100fbcdfd0258fa1a5b073d18f688c2fb3d8f9a7c59204c6777f2bbf1faeb1eb1ed",
	GenesisBlockTimestamp:     "2017-05-06T10:38:50+02:00",
}

var chainParams = defaultChainParams

const chainParamsBaseName = "chainparams.json"

// Blocks (SQLite databases) are stored as flat files in a directory
const blockchainSubdirectoryBaseName = "blocks"
const rawBlockFilenameFormat = "%s/%04x/block_%08x.db"
const rawBlockDirnameFormat = "%s/%04x"
const genesisBlockHeight = 0

var blockchainSubdirectory string


// Block is the working representation of a blockchain block
type Block struct {
	*DbBlockchainBlock
	db *sql.DB
}

// BlockKeyOp is the representation of a key op record from the blocks' _keys table.
type BlockKeyOp struct {
	op               string
	publicKeyHash    string
	publicKeyBytes   []byte
	signatureKeyHash string
	signature        []byte
	metadata         map[string]string
}

func ensureBlockchainSubdirectoryExists() {
	blockchainSubdirectory = fmt.Sprintf("%s/%s", cfg.DataDir, blockchainSubdirectoryBaseName)
	if _, err := os.Stat(blockchainSubdirectory); err != nil {
		// Probably doesn't exist, create it
		log.Println("Creating directory", blockchainSubdirectory)
		err := os.Mkdir(blockchainSubdirectory, 0700)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// Initializes the blockchain: creates database entries and the genesis block file
func blockchainInit(createDefault bool) {
	ensureBlockchainSubdirectoryExists()
	if dbGetBlockchainHeight() == -1 && createDefault {
		log.Println("Writing down the default Genesis block. Let there be light.")

		// This is basically testing the crypto code, no real purpose.
		keypair, publicKeyHash, err := cryptoGetAPrivateKey()
		if err != nil {
			log.Panicln(err)
		}
		signature, err := cryptoSignPublicKeyHash(keypair, publicKeyHash)
		if err != nil {
			log.Panicln(err)
		}
		if err = cryptoVerifyPublicKeyHashSignature(&keypair.PublicKey, publicKeyHash, signature); err != nil {
			log.Panicln(err)
		}

		/*
			// Sign the Genesis block's "previous block" hash
			genesisPrevBlockHash, err := hex.DecodeString(GenesisBlockPreviousBlockHash)
			if err != nil {
				log.Panicln(err)
			}
			signature, err = cryptoSignBytes(keypair, genesisPrevBlockHash)
			if err != nil {
				log.Panicln(err)
			}
			log.Println(GenesisBlockPreviousBlockHash)
			log.Println(hex.EncodeToString(signature))
		*/

		// Bring the genesis block into existence
		genesisBlock := MustAsset("bindata/genesis.db")
		if hashBytesToHexString(genesisBlock) != chainParams.GenesisBlockHash {
			log.Panicln("Genesis block hash unexpected:", hashBytesToHexString(genesisBlock))
		}

		/*
			// Sign the Genesis block's hash
			genesisBlockHash, err := hex.DecodeString(GenesisBlockHash)
			if err != nil {
				log.Panicln(err)
			}
			signature, err = cryptoSignBytes(keypair, genesisBlockHash)
			if err != nil {
				log.Panicln(err)
			}
			log.Println(hex.EncodeToString(signature))
		*/

		if err := blockchainEnsureBlockDir(genesisBlockHeight); err != nil {
			log.Panicln(err)
		}
		genesisBlockFilename := blockchainGetFilename(genesisBlockHeight)
		err = ioutil.WriteFile(genesisBlockFilename, genesisBlock, 0644)
		if err != nil {
			log.Panic(err)
		}
		b, err := OpenBlockFile(genesisBlockFilename)
		if err != nil {
			log.Panicln(err)
		}
		b.Height = 0
		b.TimeAccepted, err = time.Parse(time.RFC3339, chainParams.GenesisBlockTimestamp)
		if err != nil {
			log.Panicln("Error parsing genesis block timestamp", err)
		}
		b.HashSignature, err = hex.DecodeString(chainParams.GenesisBlockHashSignature)
		if err != nil {
			log.Panicln(err)
		}
		blockKeyOps, err := b.dbGetKeyOps()
		if err != nil {
			log.Panicln(err)
		}
		for _, keyOps := range blockKeyOps {
			for _, keyOp := range keyOps {
				if dbPublicKeyExists(keyOp.publicKeyHash) {
					continue
				}
				dbWritePublicKey(keyOp.publicKeyBytes, keyOp.publicKeyHash, 0)
			}
		}
		err = dbInsertBlock(b.DbBlockchainBlock)
		if err != nil {
			log.Panicln(err)
		}
	} else {
		// The chainparams file will only exist for non-default blockchains
		cpFilename := fmt.Sprintf("%s/%s", cfg.DataDir, chainParamsBaseName)
		if fileExists(cpFilename) {
			log.Println("Loading custom blockchain params from", cpFilename)
			cpJSON, err := ioutil.ReadFile(cpFilename)
			if err != nil {
				log.Fatal("Error reading chainparams file", cpFilename, err)
			}
			err = json.Unmarshal(cpJSON, &chainParams)
			if err != nil {
				log.Fatal("Error decoding chainparams file", cpFilename, err)
			}
			peers := dbGetSavedPeers()
			for _, peer := range chainParams.BootstrapPeers {
				_, ok := peers[peer]
				if !ok {
					dbSavePeer(peer)
				}
			}
		} else {
			log.Println("Using default blockchain params")
		}
		log.Println("P2P peers:", dbGetSavedPeers())
	}
	err := blockchainVerifyEverything()
	if err != nil {
		log.Fatalf("blockchainVerifyEverything: %v", err)
	}
}

// Verifies the entire blockchain to see if there are errors.
// TODO: Dynamic adding and revoking of key is not yet checked
func blockchainVerifyEverything() error {
	if cfg.faster {
		log.Println("Skipping blockchain consistency checks")
		return nil
	}
	log.Println("Verifying all the blocks (use --faster to skip)...")
	maxHeight := dbGetBlockchainHeight()
	for height := 0; height <= maxHeight; height++ {
		if height > 0 && height%1000 == 0 {
			log.Println("Verifying block", height)
		}
		if err := blockchainEnsureBlockDir(height); err != nil {
			return err
		}
		blockFilename := blockchainGetFilename(height)
		fileHash, err := hashFileToHexString(blockFilename)
		if err != nil {
			return fmt.Errorf("block %d: %v", height, err)
		}
		dbb, err := dbGetBlockByHeight(height)
		if err != nil {
			return fmt.Errorf("block %d: %v", height, err)
		}
		if fileHash != dbb.Hash {
			return fmt.Errorf("block %d: file hash %s doesn't match db hash %s", height, fileHash, dbb.Hash)
		}
		if height == 0 && fileHash != chainParams.GenesisBlockHash {
			return fmt.Errorf("block %d: it's supposed to be the genesis block but its hash doesn't match %s",
				height, chainParams.GenesisBlockHash)
		}
		dbpk, err := dbGetPublicKey(dbb.SignaturePublicKeyHash)
		if err != nil {
			return fmt.Errorf("block %d: error getting public key %s", height, dbb.SignaturePublicKeyHash)
		}
		creatorPublicKey, err := cryptoDecodePublicKeyBytes(dbpk.publicKeyBytes)
		if err != nil {
			return fmt.Errorf("block %d: cannot decode public key %s", height, dbb.SignaturePublicKeyHash)
		}
		hashBytes, err := hex.DecodeString(dbb.Hash)
		if err != nil {
			return fmt.Errorf("block %d: cannot decode hash %s", height, dbb.Hash)
		}
		err = cryptoVerifyBytes(creatorPublicKey, hashBytes, dbb.HashSignature)
		if err != nil {
			log.Println(creatorPublicKey, hashBytes, dbb.HashSignature)
			return fmt.Errorf("block %d: block hash signature is invalid (%v)", height, err)
		}
		previousHashBytes, err := hex.DecodeString(dbb.PreviousBlockHash)
		if err != nil {
			return fmt.Errorf("block %d: cannot decode previous block hash %s", height, dbb.PreviousBlockHash)
		}
		err = cryptoVerifyBytes(creatorPublicKey, previousHashBytes, dbb.PreviousBlockHashSignature)
		if err != nil {
			return fmt.Errorf("block %d: previous block hash signature is invalid (%v)", height, err)
		}
		b, err := OpenBlockByHeight(height)
		if err != nil {
			return fmt.Errorf("block %d: cannot open block db file: %v", height, err)
		}
		blockKeyOps, err := b.dbGetKeyOps()
		if err != nil {
			if err := b.Close(); err != nil {
				panic(err)
			}
			return fmt.Errorf("block %d: cannot get key ops: %v", height, err)
		}
		if err = b.Close(); err != nil {
			panic(err)
		}
		Q := QuorumForHeight(height)
		for keyOpKeyHash, keyOps := range blockKeyOps {
			if len(keyOps) != Q {
				return fmt.Errorf("block %d: key ops for %s don't have quorum: %d vs Q=%d",
					height, keyOpKeyHash, len(keyOps), Q)
			}
			op := keyOps[0].op
			for _, kop := range keyOps {
				if kop.op != op {
					return fmt.Errorf("block %d: key ops for %s don't match: %s vs %s",
						height, keyOpKeyHash, kop.op, op)
				}
				dbSigningKey, err := dbGetPublicKey(kop.signatureKeyHash)
				if err != nil {
					return fmt.Errorf("block %d: cannot get public key %s from main db", height, kop.signatureKeyHash)
				}
				signingKey, err := cryptoDecodePublicKeyBytes(dbSigningKey.publicKeyBytes)
				if err != nil {
					return fmt.Errorf("block %d: cannot decode public key %s", height, dbSigningKey.publicKeyHash)
				}
				if err = cryptoVerifyPublicKeyHashSignature(signingKey, kop.publicKeyHash, kop.signature); err != nil {
					return fmt.Errorf("block %d: key op signature invalid for signer %s: %v", height, kop.signatureKeyHash, err)
				}
			}
		}
	}
	return nil
}

// Checks if a new block can be accepted to extend the blockchain
func checkAcceptBlock(blk *Block) (int, error) {
	// Step 1: Does the block fit, i.e. does it extend the chain?
	if blk.Version != CurrentBlockVersion {
		return 0, fmt.Errorf("Unsupported block version: %d", blk.Version)
	}
	prevBlk, err := dbGetBlock(blk.PreviousBlockHash)
	if err != nil {
		return 0, fmt.Errorf("Cannot find previous block %s: %v", blk.PreviousBlockHash, err)
	}
	thisBlockHeight := prevBlk.Height + 1
	if _, err = dbGetBlockByHeight(thisBlockHeight); err == nil {
		return 0, fmt.Errorf("The block to accept would replace an existing block, and this is not supported yet (height=%d)", prevBlk.Height+1)
	}
	// Step 2: Is the block signed by a valid signatory?
	signatoryPubKey, err := dbGetPublicKey(blk.SignaturePublicKeyHash)
	if err != nil {
		return 0, fmt.Errorf("Cannot find an accepted public key %s signing the block", blk.SignaturePublicKeyHash)
	}
	if signatoryPubKey.isRevoked {
		return 0, fmt.Errorf("The public key %s signing the block is revoked on %v", blk.SignaturePublicKeyHash, signatoryPubKey.timeRevoked)
	}
	sigPubKey, err := cryptoDecodePublicKeyBytes(signatoryPubKey.publicKeyBytes)
	if err != nil {
		return 0, fmt.Errorf("Cannot decode public key %s: %v", blk.SignaturePublicKeyHash, err)
	}
	err = cryptoVerifyHexBytes(sigPubKey, blk.PreviousBlockHash, blk.PreviousBlockHashSignature)
	if err != nil {
		return 0, fmt.Errorf("Verification of previous block hash has failed: %v", err)
	}
	err = cryptoVerifyHexBytes(sigPubKey, blk.Hash, blk.HashSignature)
	if err != nil {
		return 0, fmt.Errorf("Verification of block hash has failed: %v", err)
	}
	allKeyOps, err := blk.dbGetKeyOps()
	if err != nil {
		return 0, err
	}
	targetQuorum := QuorumForHeight(thisBlockHeight)
	for key, keyOps := range allKeyOps {
		if len(keyOps) < targetQuorum {
			return 0, fmt.Errorf("Quorum of %d not met for key ops on key %s", targetQuorum, key)
		}
		for _, keyOp := range keyOps {
			signatoryPubKey, err = dbGetPublicKey(keyOp.signatureKeyHash)
			if err != nil {
				return 0, fmt.Errorf("Error retrieving supposedly key op signatory %s", keyOp.signatureKeyHash)
			}
			sigPubKey, err := cryptoDecodePublicKeyBytes(signatoryPubKey.publicKeyBytes)
			if err != nil {
				return 0, fmt.Errorf("Cannot decode public key %s: %v", signatoryPubKey.publicKeyHash, err)
			}
			err = cryptoVerifyPublicKeyHashSignature(sigPubKey, key, keyOp.signature)
			if err != nil {
				return 0, fmt.Errorf("Failed verification of key op for %s by %s", key, keyOp.signatureKeyHash)
			}
		}
		// At this point, all required signatures have been verified
		if keyOps[0].op == "A" {
			// Add the key to the list of valid signatories. But first, check if it already exists.
			_, err := dbGetPublicKey(key)
			if err == nil {
				return 0, fmt.Errorf("Attempt to add an already existing key to the list of signatores")
			}
			dbWritePublicKey(keyOps[0].publicKeyBytes, key, thisBlockHeight)
		} else if keyOps[0].op == "R" {
			// Revoke the key. But first, check if it's already revoked.
			dbpk, err := dbGetPublicKey(key)
			if err != nil {
				return 0, fmt.Errorf("Cannot retrieve key to revoke: %s", key)
			}
			if dbpk.isRevoked {
				return 0, fmt.Errorf("Attempt to revoke a key which is already revoked: %s", key)
			}
			dbRevokePublicKey(key)
		} else {
			return 0, fmt.Errorf("Invalid key op: %s", keyOps[0].op)
		}
	}
	// Everything's ok, the block is ok to import.
	return thisBlockHeight, nil
}

// QuorumForHeight calculates the required key op quorum for the given block height
func QuorumForHeight(h int) int {
	if h < 149 {
		return 1
	}
	return int(math.Log(float64(h)) * 2)
}

// Formats the block height into a blockchain file (SQLite database) filename
func blockchainGetFilename(h int) string {
	return fmt.Sprintf(rawBlockFilenameFormat, blockchainSubdirectory, h/65536, h)
}

func blockchainEnsureBlockDir(h int) error {
	dirName := fmt.Sprintf(rawBlockDirnameFormat, blockchainSubdirectory, h/65536)
	return os.MkdirAll(dirName, 0755)
}

// OpenBlockByHeight opens a block stored in the blockchain at the given height
func OpenBlockByHeight(height int) (*Block, error) {
	b := Block{DbBlockchainBlock: &DbBlockchainBlock{Height: height}}
	if err := blockchainEnsureBlockDir(height); err != nil {
		return nil, err
	}
	blockFilename := blockchainGetFilename(height)
	hash, err := hashFileToHexString(blockFilename)
	if err != nil {
		return nil, err
	}
	dbb, err := dbGetBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	if hash != dbb.Hash {
		return nil, fmt.Errorf("Recorded block hash doesn't match actual: %s vs %s", dbb.Hash, hash)
	}
	b.DbBlockchainBlock = dbb
	b.db, err = dbOpen(blockFilename, true)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// OpenBlockFile reads block metadata from the given database file.
// Note that it will not fill-in all the fields. Notable, height is not stored in the block db's metadata.
func OpenBlockFile(fileName string) (*Block, error) {
	hash, err := hashFileToHexString(fileName)
	if err != nil {
		return nil, err
	}
	db, err := dbOpen(fileName, true)
	if err != nil {
		return nil, err
	}
	b := Block{DbBlockchainBlock: &DbBlockchainBlock{Hash: hash}, db: db}
	if b.Version, err = b.dbGetMetaInt("Version"); err != nil {
		return nil, err
	}
	if b.PreviousBlockHash, err = b.dbGetMetaString("PreviousBlockHash"); err != nil {
		return nil, err
	}
	if b.SignaturePublicKeyHash, err = b.dbGetMetaString("CreatorPublicKey"); err != nil {
		return nil, err
	}
	if b.PreviousBlockHashSignature, err = b.dbGetMetaHexBytes("PreviousBlockHashSignature"); err != nil {
		return nil, err
	}
	if b.TimeAccepted, err = b.dbGetMetaTime("Timestamp"); err != nil {
		st, err := os.Stat(fileName)
		if err != nil {
			return nil, err
		}
		b.TimeAccepted = st.ModTime()
	}
	return &b, nil
}

func (b *Block) Close() error {
	return b.db.Close()
}

// Returns an integer value from the _meta table within the block
func (b *Block) dbGetMetaInt(key string) (int, error) {
	var value string
	if err := b.db.QueryRow("SELECT value FROM _meta WHERE key=?", key).Scan(&value); err != nil {
		return -1, err
	}
	return strconv.Atoi(value)
}

// Returns a timestamp value from the _meta table within the block
func (b *Block) dbGetMetaTime(key string) (time.Time, error) {
	var value string
	if err := b.db.QueryRow("SELECT value FROM _meta WHERE key=?", key).Scan(&value); err != nil {
		return time.Time{}, err
	}
	t, err := time.Parse(time.RFC3339, value)
	return t, err
}

// Returns a string value from the _meta table within the block
func (b *Block) dbGetMetaString(key string) (string, error) {
	var value string
	if err := b.db.QueryRow("SELECT value FROM _meta WHERE key=?", key).Scan(&value); err != nil {
		return "", err
	}
	return value, nil
}

// Returns a byte blob value from the _meta table within the block (stored in the db as a hex string)
func (b *Block) dbGetMetaHexBytes(key string) ([]byte, error) {
	var value string
	if err := b.db.QueryRow("SELECT value FROM _meta WHERE key=?", key).Scan(&value); err != nil {
		return nil, err
	}
	return hex.DecodeString(value)
}

// Returns a map of key operations stored in the block. Map keys are public key hashes, values are lists of ops.
func (b *Block) dbGetKeyOps() (map[string][]BlockKeyOp, error) {
	var count int
	if err := b.db.QueryRow("SELECT COUNT(*) FROM _keys").Scan(&count); err != nil {
		log.Println("Error reading db _keys")
		return nil, err
	}
	keyOps := make(map[string][]BlockKeyOp)
	rows, err := b.db.Query("SELECT op, pubkey_hash, pubkey, sigkey_hash, signature, COALESCE(metadata, '') FROM _keys")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var publicKeyHex string
		var signatureHex string
		var metadataJSON string
		var keyOp BlockKeyOp
		if err = rows.Scan(&keyOp.op, &keyOp.publicKeyHash, &publicKeyHex, &keyOp.signatureKeyHash, &signatureHex, &metadataJSON); err != nil {
			return nil, err
		}
		if keyOp.publicKeyBytes, err = hex.DecodeString(publicKeyHex); err != nil {
			return nil, err
		}
		publicKey, err := cryptoDecodePublicKeyBytes(keyOp.publicKeyBytes)
		if err != nil {
			return nil, err
		}
		if keyOp.publicKeyHash != cryptoMustGetPublicKeyHash(publicKey) {
			return nil, fmt.Errorf("Public key hash doesn't match for %s", keyOp.publicKeyHash)
		}
		if keyOp.signature, err = hex.DecodeString(signatureHex); err != nil {
			return nil, err
		}
		if metadataJSON != "" {
			if err = json.Unmarshal([]byte(metadataJSON), &keyOp.metadata); err != nil {
				return nil, err
			}
		}
		if _, ok := keyOps[keyOp.publicKeyHash]; ok {
			keyOps[keyOp.publicKeyHash] = append(keyOps[keyOp.publicKeyHash], keyOp)
		} else {
			keyOps[keyOp.publicKeyHash] = make([]BlockKeyOp, 1)
			keyOps[keyOp.publicKeyHash][0] = keyOp
		}
	}
	for hash, oneKeyOps := range keyOps {
		for _, keyOp := range oneKeyOps {
			if keyOp.op != oneKeyOps[0].op {
				return nil, fmt.Errorf("Mixed key ops for a single public key %s", hash)
			}
		}
	}
	return keyOps, nil
}

// Ensures special metadata tables exist in a SQLite database
func dbEnsureBlockchainTables(db *sql.DB) {
	if !dbTableExists(db, "_meta") {
		_, err := db.Exec(metaTableCreate)
		if err != nil {
			log.Fatal(err)
		}
	}
	if !dbTableExists(db, "_keys") {
		_, err := db.Exec(keysTableCreate)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Stores a key-value pair into the _meta table in the SQLite database
func dbSetMetaString(db *sql.DB, key string, value string) error {
	_, err := db.Exec("INSERT OR REPLACE INTO _meta(key, value) VALUES (?, ?)", key, value)
	return err
}

// Stores a key-value pair into the _meta table in the SQLite database
func dbSetMetaInt(db *sql.DB, key string, value int) error {
	_, err := db.Exec("INSERT OR REPLACE INTO _meta(key, value) VALUES (?, ?)", key, value)
	return err
}

// Copies a given file to the blockchain directory and names it as a block with the given height
func blockchainCopyFile(fn string, height int) error {
	if err := blockchainEnsureBlockDir(height); err != nil {
		return err
	}
	blockFilename := blockchainGetFilename(height)
	in, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer func() {
		err = in.Close()
		if err != nil {
			log.Printf("blockchainCopyFile in.Close: %v", err)
		}
	}()
	out, err := os.Create(blockFilename)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
