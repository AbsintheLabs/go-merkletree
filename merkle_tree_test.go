// MIT License
//
// Copyright (c) 2022 Tommy TIAN
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package merkletree

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/txaty/go-merkletree/mock"
)

const benchSize = 10000

func dataBlocks(num int) []DataBlock {
	blocks := make([]DataBlock, num)
	for i := 0; i < num; i++ {
		block := &mock.DataBlock{
			Data: make([]byte, 100),
		}
		if _, err := rand.Read(block.Data); err != nil {
			panic(err)
		}
		blocks[i] = block
	}
	return blocks
}

func TestMerkleTreeNew_proofGen(t *testing.T) {
	dummyDataBlocks := []DataBlock{
		&mock.DataBlock{
			Data: []byte("dummy_data_0"),
		},
		&mock.DataBlock{
			Data: []byte("dummy_data_1"),
		},
		&mock.DataBlock{
			Data: []byte("dummy_data_2"),
		},
	}
	dummyHashList := make([][]byte, 3)
	var err error
	for i := 0; i < 3; i++ {
		dataByte, err := dummyDataBlocks[i].Serialize()
		if err != nil {
			t.Fatal(err)
		}
		dummyHashList[i], err = DefaultHashFunc(dataByte)
		if err != nil {
			t.Fatal(err)
		}
	}
	twoDummyRoot, err := DefaultHashFunc(
		append(dummyHashList[0], dummyHashList[1]...),
	)
	if err != nil {
		t.Fatal(err)
	}
	leftHash, err := DefaultHashFunc(
		append(dummyHashList[0], dummyHashList[1]...),
	)
	if err != nil {
		t.Fatal(err)
	}
	rightHash, err := DefaultHashFunc(
		append(dummyHashList[2], dummyHashList[2]...),
	)
	if err != nil {
		t.Fatal(err)
	}
	threeDummyRoot, err := DefaultHashFunc(append(leftHash, rightHash...))
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		blocks []DataBlock
		config *Config
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantRoot []byte
	}{
		{
			name: "test_0",
			args: args{
				blocks: dataBlocks(0),
			},
			wantErr: true,
		},
		{
			name: "test_1",
			args: args{
				blocks: dataBlocks(1),
			},
			wantErr: true,
		},
		{
			name: "test_2",
			args: args{
				blocks: []DataBlock{dummyDataBlocks[0], dummyDataBlocks[1]},
			},
			wantErr:  false,
			wantRoot: twoDummyRoot,
		},
		{
			name: "test_3",
			args: args{
				blocks: dummyDataBlocks,
			},
			wantErr:  false,
			wantRoot: threeDummyRoot,
		},
		{
			name: "test_8",
			args: args{
				blocks: dataBlocks(8),
			},
			wantErr: false,
		},
		{
			name: "test_5",
			args: args{
				blocks: dataBlocks(5),
			},
			wantErr: false,
		},
		{
			name: "test_1000",
			args: args{
				blocks: dataBlocks(1000),
			},
			wantErr: false,
		},
		{
			name: "test_100_parallel_4",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   4,
				},
			},
			wantErr: false,
		},
		{
			name: "test_10_parallel_32",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   32,
				},
			},
			wantErr: false,
		},
		{
			name: "test_100_parallel_no_specify_num_of_routines",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					RunInParallel: true,
				},
			},
			wantErr: false,
		},
		{
			name: "test_8_sorted",
			args: args{
				blocks: dataBlocks(8),
				config: &Config{
					SortSiblingPairs: true,
				},
			},
			wantErr: false,
		},
		{
			name: "test_hash_func_error",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					HashFunc: func([]byte) ([]byte, error) {
						return nil, fmt.Errorf("hash func error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test_hash_func_error_parallel",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					HashFunc: func([]byte) ([]byte, error) {
						return nil, fmt.Errorf("hash func error")
					},
					RunInParallel: true,
				},
			},
			wantErr: true,
		},
		{
			name: "test_100_disable_leaf_hashing",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					DisableLeafHashing: true,
				},
			},
			wantErr: false,
		},
		{
			name: "test_100_disable_leaf_hashing_parallel_4",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					DisableLeafHashing: true,
					RunInParallel:      true,
					NumRoutines:        4,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid_mode",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					Mode: 5,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt, err := New(tt.args.config, tt.args.blocks)
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.wantRoot == nil {
				for idx, block := range tt.args.blocks {
					if ok, _ := mt.Verify(block, mt.Proofs[idx]); !ok {
						t.Errorf("proof verification failed")
						return
					}
				}
			} else {
				if !bytes.Equal(mt.Root, tt.wantRoot) {
					t.Errorf("root mismatch, got %x, want %x", mt.Root, tt.wantRoot)
					return
				}
			}
		})
	}
}

func TestMerkleTreeNew_buildTree(t *testing.T) {
	type args struct {
		blocks []DataBlock
		config *Config
	}
	tests := []struct {
		name           string
		args           args
		checkingConfig *Config
		wantErr        bool
	}{
		{
			name: "test_build_tree_2",
			args: args{
				blocks: dataBlocks(2),
				config: &Config{
					Mode: ModeTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_5",
			args: args{
				blocks: dataBlocks(5),
				config: &Config{
					Mode: ModeTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_8",
			args: args{
				blocks: dataBlocks(8),
				config: &Config{
					Mode: ModeTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_1000",
			args: args{
				blocks: dataBlocks(1000),
				config: &Config{
					Mode: ModeTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_hash_func_error",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					HashFunc: func([]byte) ([]byte, error) {
						return nil, fmt.Errorf("hash func error")
					},
					Mode: ModeTreeBuild,
				},
			},
			wantErr: true,
		},
		{
			name: "test_disable_leaf_hashing",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					DisableLeafHashing: true,
					Mode:               ModeTreeBuild,
				},
			},
			checkingConfig: &Config{
				DisableLeafHashing: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.args.config, tt.args.blocks)
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			m1, err := New(tt.checkingConfig, tt.args.blocks)
			if err != nil {
				t.Errorf("test setup error %v", err)
				return
			}
			if !tt.wantErr && !bytes.Equal(m.Root, m1.Root) && !tt.wantErr {
				fmt.Println("m", m.Root)
				fmt.Println("m1", m1.Root)
				t.Errorf("tree generated is wrong")
				return
			}
		})
	}
}

func TestMerkleTreeNew_treeBuildParallel(t *testing.T) {
	type args struct {
		blocks []DataBlock
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_build_tree_parallel_2",
			args: args{
				blocks: dataBlocks(2),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   4,
					Mode:          ModeTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_parallel_4",
			args: args{
				blocks: dataBlocks(4),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   4,
					Mode:          ModeTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_parallel_5",
			args: args{
				blocks: dataBlocks(5),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   4,
					Mode:          ModeTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_parallel_8",
			args: args{
				blocks: dataBlocks(8),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   4,
					Mode:          ModeTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_parallel_8_32",
			args: args{
				blocks: dataBlocks(8),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   32,
					Mode:          ModeTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_hash_func_error_parallel",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					HashFunc: func([]byte) ([]byte, error) {
						return nil, fmt.Errorf("hash func error")
					},
					RunInParallel: true,
					Mode:          ModeTreeBuild,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.args.config, tt.args.blocks)
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			m1, err := New(nil, tt.args.blocks)
			if err != nil {
				t.Errorf("test setup error %v", err)
				return
			}
			if !tt.wantErr && !bytes.Equal(m.Root, m1.Root) && !tt.wantErr {
				fmt.Println("m", m.Root)
				fmt.Println("m1", m1.Root)
				t.Errorf("tree generated is wrong")
				return
			}
		})
	}
}

func TestMerkleTreeNew_proofGenAndTreeBuild(t *testing.T) {
	type args struct {
		blocks []DataBlock
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_build_tree_proof_2",
			args: args{
				blocks: dataBlocks(2),
				config: &Config{
					Mode: ModeProofGenAndTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_proof_4",
			args: args{
				blocks: dataBlocks(4),
				config: &Config{
					Mode: ModeProofGenAndTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_proof_5",
			args: args{
				blocks: dataBlocks(5),
				config: &Config{
					Mode: ModeProofGenAndTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_proof_8",
			args: args{
				blocks: dataBlocks(8),
				config: &Config{
					Mode: ModeProofGenAndTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_proof_9",
			args: args{
				blocks: dataBlocks(9),
				config: &Config{
					Mode: ModeProofGenAndTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_hash_func_error",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					HashFunc: func([]byte) ([]byte, error) {
						return nil, fmt.Errorf("hash func error")
					},
					Mode: ModeProofGenAndTreeBuild,
				},
			},
			wantErr: true,
		},
		{
			name: "test_tree_build_hash_func_error",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					HashFunc: func(block []byte) ([]byte, error) {
						if len(block) == 64 {
							return nil, fmt.Errorf("hash func error")
						}
						sha256Func := sha256.New()
						sha256Func.Write(block)
						return sha256Func.Sum(nil), nil
					},
					Mode: ModeProofGenAndTreeBuild,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.args.config, tt.args.blocks)
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			m1, err := New(nil, tt.args.blocks)
			if err != nil {
				t.Errorf("test setup error %v", err)
				return
			}
			for i := 0; i < len(tt.args.blocks); i++ {
				if !reflect.DeepEqual(m.Proofs[i], m1.Proofs[i]) {
					t.Errorf("proofs generated are wrong for block %d", i)
					return
				}
			}
		})
	}
}

func TestMerkleTreeNew_proofGenAndTreeBuildParallel(t *testing.T) {
	type args struct {
		blocks []DataBlock
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_build_tree_proof_parallel_2",
			args: args{
				blocks: dataBlocks(2),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   4,
					Mode:          ModeProofGenAndTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_proof_parallel_4",
			args: args{
				blocks: dataBlocks(4),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   4,
					Mode:          ModeProofGenAndTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_proof_parallel_5",
			args: args{
				blocks: dataBlocks(5),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   4,
					Mode:          ModeProofGenAndTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_build_tree_proof_parallel_8",
			args: args{
				blocks: dataBlocks(8),
				config: &Config{
					RunInParallel: true,
					NumRoutines:   4,
					Mode:          ModeProofGenAndTreeBuild,
				},
			},
			wantErr: false,
		},
		{
			name: "test_hash_func_error",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					HashFunc: func([]byte) ([]byte, error) {
						return nil, fmt.Errorf("hash func error")
					},
					Mode:          ModeProofGenAndTreeBuild,
					RunInParallel: true,
				},
			},
			wantErr: true,
		},
		{
			name: "test_tree_build_hash_func_error",
			args: args{
				blocks: dataBlocks(100),
				config: &Config{
					HashFunc: func(block []byte) ([]byte, error) {
						if len(block) == 64 {
							return nil, fmt.Errorf("hash func error")
						}
						sha256Func := sha256.New()
						sha256Func.Write(block)
						return sha256Func.Sum(nil), nil
					},
					Mode:          ModeProofGenAndTreeBuild,
					RunInParallel: true,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.args.config, tt.args.blocks)
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			m1, err := New(nil, tt.args.blocks)
			if err != nil {
				t.Errorf("test setup error %v", err)
				return
			}
			for i := 0; i < len(tt.args.blocks); i++ {
				if !reflect.DeepEqual(m.Proofs[i], m1.Proofs[i]) {
					t.Errorf("proofs generated are wrong for block %d", i)
					return
				}
			}
		})
	}
}

func verifySetup(size int) (*MerkleTree, []DataBlock) {
	blocks := dataBlocks(size)
	m, err := New(nil, blocks)
	if err != nil {
		panic(err)
	}
	return m, blocks
}

func verifySetupParallel(size int) (*MerkleTree, []DataBlock) {
	blocks := dataBlocks(size)
	m, err := New(&Config{
		RunInParallel: true,
		NumRoutines:   1,
	}, blocks)
	if err != nil {
		panic(err)
	}
	return m, blocks
}

func TestMerkleTree_Verify(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(int) (*MerkleTree, []DataBlock)
		blockSize int
		want      bool
		wantErr   bool
	}{
		{
			name:      "test_2",
			setupFunc: verifySetup,
			blockSize: 2,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_3",
			setupFunc: verifySetup,
			blockSize: 3,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_4",
			setupFunc: verifySetup,
			blockSize: 4,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_5",
			setupFunc: verifySetup,
			blockSize: 5,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_6",
			setupFunc: verifySetup,
			blockSize: 6,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_8",
			setupFunc: verifySetup,
			blockSize: 8,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_9",
			setupFunc: verifySetup,
			blockSize: 9,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_1001",
			setupFunc: verifySetup,
			blockSize: 1001,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_2_parallel",
			setupFunc: verifySetupParallel,
			blockSize: 2,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_4_parallel",
			setupFunc: verifySetupParallel,
			blockSize: 4,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_64_parallel",
			setupFunc: verifySetupParallel,
			blockSize: 64,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "test_1001_parallel",
			setupFunc: verifySetupParallel,
			blockSize: 1001,
			want:      true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, blocks := tt.setupFunc(tt.blockSize)
			for i := 0; i < tt.blockSize; i++ {
				got, err := m.Verify(blocks[i], m.Proofs[i])
				if (err != nil) != tt.wantErr {
					t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("Verify() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMerkleTree_Proof(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	tests := []struct {
		name        string
		config      *Config
		mock        func()
		blocks      []DataBlock
		proofBlocks []DataBlock
		wantErr     bool
	}{
		{
			name:   "test_2",
			config: &Config{Mode: ModeTreeBuild},
			blocks: dataBlocks(2),
		},
		{
			name:   "test_4",
			config: &Config{Mode: ModeTreeBuild},
			blocks: dataBlocks(4),
		},
		{
			name:   "test_5",
			config: &Config{Mode: ModeTreeBuild},
			blocks: dataBlocks(5),
		},
		{
			name:    "test_wrong_mode",
			config:  &Config{Mode: ModeProofGen},
			blocks:  dataBlocks(5),
			wantErr: true,
		},
		{
			name:   "test_wrong_blocks",
			config: &Config{Mode: ModeTreeBuild},
			blocks: dataBlocks(5),
			proofBlocks: []DataBlock{
				&mock.DataBlock{
					Data: []byte("test_wrong_blocks"),
				},
			},
			wantErr: true,
		},
		{
			name:   "test_data_block_serialize_error",
			config: &Config{Mode: ModeTreeBuild},
			mock: func() {
				patches.ApplyMethod(reflect.TypeOf(&mock.DataBlock{}), "Serialize",
					func(*mock.DataBlock) ([]byte, error) {
						return nil, errors.New("data block serialize error")
					})
			},
			blocks:  dataBlocks(5),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m1, err := New(nil, tt.blocks)
			if err != nil {
				t.Errorf("m1 New() error = %v", err)
				return
			}
			m2, err := New(tt.config, tt.blocks)
			if err != nil {
				t.Errorf("m2 New() error = %v", err)
				return
			}
			if tt.proofBlocks == nil {
				tt.proofBlocks = tt.blocks
			}
			if tt.mock != nil {
				tt.mock()
			}
			defer patches.Reset()
			for idx, block := range tt.proofBlocks {
				got, err := m2.Proof(block)
				if (err != nil) != tt.wantErr {
					t.Errorf("Proof() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErr {
					return
				}
				if !reflect.DeepEqual(got, m1.Proofs[idx]) && !tt.wantErr {
					t.Errorf("Proof() %d got = %v, want %v", idx, got, m1.Proofs[idx])
					return
				}
			}
		})
	}
}

func mockHashFunc(data []byte) ([]byte, error) {
	sha256Func := sha256.New()
	sha256Func.Write(data)
	return sha256Func.Sum(nil), nil
}

func TestMerkleTree_proofGen(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	type args struct {
		config *Config
		blocks []DataBlock
	}
	tests := []struct {
		name    string
		args    args
		mock    func()
		wantErr bool
	}{
		{
			name: "test_hash_func_err",
			args: args{
				config: &Config{
					HashFunc: mockHashFunc,
				},
				blocks: dataBlocks(5),
			},
			mock: func() {
				patches.ApplyFunc(mockHashFunc,
					func([]byte) ([]byte, error) {
						return nil, errors.New("test_hash_func_err")
					})
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.args.config, tt.args.blocks)
			if err != nil {
				t.Errorf("New() error = %v", err)
				return
			}
			if tt.mock != nil {
				tt.mock()
			}
			defer patches.Reset()
			if err := m.proofGen(); (err != nil) != tt.wantErr {
				t.Errorf("proofGen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	blocks := dataBlocks(5)
	m, err := New(nil, blocks)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	type args struct {
		dataBlock DataBlock
		proof     *Proof
		root      []byte
		config    *Config
	}
	tests := []struct {
		name    string
		args    args
		mock    func()
		want    bool
		wantErr bool
	}{
		{
			name: "test_ok",
			args: args{
				dataBlock: blocks[0],
				proof:     m.Proofs[0],
				root:      m.Root,
				config: &Config{
					HashFunc: m.HashFunc,
					concatFunc: func(left, right []byte) []byte {
						return append(left, right...)
					},
				},
			},
			want: true,
		},
		{
			name: "test_config_nil",
			args: args{
				dataBlock: blocks[0],
				proof:     m.Proofs[0],
				root:      m.Root,
			},
			want: true,
		},
		{
			name: "test_wrong_root",
			args: args{
				dataBlock: blocks[0],
				proof:     m.Proofs[0],
				root:      []byte("test_wrong_root"),
				config: &Config{
					HashFunc: m.HashFunc,
					concatFunc: func(left, right []byte) []byte {
						return append(left, right...)
					},
				},
			},
			want: false,
		},
		{
			name: "test_wrong_hash_func",
			args: args{
				dataBlock: blocks[0],
				proof:     m.Proofs[0],
				root:      m.Root,
				config: &Config{
					HashFunc: func([]byte) ([]byte, error) { return []byte("test_wrong_hash_hash"), nil },
					concatFunc: func(left, right []byte) []byte {
						return append(left, right...)
					},
				},
			},
			want: false,
		},
		{
			name: "test_proof_nil",
			args: args{
				dataBlock: blocks[0],
				proof:     nil,
				root:      m.Root,
				config: &Config{
					HashFunc: m.HashFunc,
					concatFunc: func(left, right []byte) []byte {
						return append(left, right...)
					},
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "test_data_block_nil",
			args: args{
				dataBlock: nil,
				proof:     m.Proofs[0],
				root:      m.Root,
				config: &Config{
					HashFunc: m.HashFunc,
					concatFunc: func(left, right []byte) []byte {
						return append(left, right...)
					},
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "test_hash_func_nil",
			args: args{
				dataBlock: blocks[0],
				proof:     m.Proofs[0],
				root:      m.Root,
				config: &Config{
					HashFunc: nil,
					concatFunc: func(left, right []byte) []byte {
						return append(left, right...)
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test_hash_func_err",
			args: args{
				dataBlock: blocks[0],
				proof:     m.Proofs[0],
				root:      m.Root,
				config: &Config{
					HashFunc: func([]byte) ([]byte, error) {
						return nil, errors.New("test_hash_func_err")
					},
					concatFunc: func(left, right []byte) []byte {
						return append(left, right...)
					},
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "data_block_serialize_err",
			args: args{
				dataBlock: blocks[0],
				proof:     m.Proofs[0],
				root:      m.Root,
				config: &Config{
					HashFunc: m.HashFunc,
					concatFunc: func(left, right []byte) []byte {
						return append(left, right...)
					},
				},
			},
			mock: func() {
				patches.ApplyMethod(reflect.TypeOf(&mock.DataBlock{}), "Serialize",
					func(m *mock.DataBlock) ([]byte, error) {
						return nil, errors.New("test_data_block_serialize_err")
					})
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mock != nil {
				tt.mock()
			}
			defer patches.Reset()
			got, err := Verify(tt.args.dataBlock, tt.args.proof, tt.args.root, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_layeredHashHandler(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	type args struct {
		arg poolWorkerArgs
	}
	mt, err := New(nil, dataBlocks(5))
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}
	mt.HashFunc = func([]byte) ([]byte, error) {
		return nil, errors.New("test_hash_func_err")
	}
	tests := []struct {
		mock    func()
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_hash_func_err",
			args: args{
				arg: poolWorkerArgs{
					mt:         mt,
					byteField1: [][]byte{[]byte("test_buf1"), []byte("test_buf1")},
					byteField2: [][]byte{[]byte("test_buf2")},
					intField2:  2,
					intField3:  2,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mock != nil {
				tt.mock()
			}
			defer patches.Reset()
			if err := proofGenHandler(tt.args.arg); (err != nil) != tt.wantErr {
				t.Errorf("proofGenHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkMerkleTreeNew(b *testing.B) {
	testCases := dataBlocks(benchSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := New(nil, testCases)
		if err != nil {
			b.Errorf("Build() error = %v", err)
		}
	}
}

func BenchmarkMerkleTreeNewParallel(b *testing.B) {
	config := &Config{
		RunInParallel: true,
	}
	testCases := dataBlocks(benchSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := New(config, testCases)
		if err != nil {
			b.Errorf("Build() error = %v", err)
		}
	}
}

func BenchmarkMerkleTreeBuild(b *testing.B) {
	testCases := dataBlocks(benchSize)
	config := &Config{
		Mode: ModeTreeBuild,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := New(config, testCases)
		if err != nil {
			b.Errorf("Build() error = %v", err)
		}
	}
}

func BenchmarkMerkleTreeBuildParallel(b *testing.B) {
	config := &Config{
		Mode:          ModeTreeBuild,
		RunInParallel: true,
	}
	testCases := dataBlocks(benchSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := New(config, testCases)
		if err != nil {
			b.Errorf("Build() error = %v", err)
		}
	}
}
