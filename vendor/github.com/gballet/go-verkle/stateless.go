// This is free and unencumbered software released into the public domain.
//
// Anyone is free to copy, modify, publish, use, compile, sell, or
// distribute this software, either in source code form or as a compiled
// binary, for any purpose, commercial or non-commercial, and by any
// means.
//
// In jurisdictions that recognize copyright laws, the author or authors
// of this software dedicate any and all copyright interest in the
// software to the public domain. We make this dedication for the benefit
// of the public at large and to the detriment of our heirs and
// successors. We intend this dedication to be an overt act of
// relinquishment in perpetuity of all present and future rights to this
// software under copyright law.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
//
// For more information, please refer to <https://unlicense.org>

package verkle

import (
	"bytes"
	"errors"
	"fmt"
)

// StatelessNode represents a node for execution in a stateless context,
// i.e. that its children/values are not all known. It can represent both
// an InternalNode or a LeafNode.
type StatelessNode struct {
	// List of available child nodes of this internal node,
	// nil if this is an extension node.
	children map[byte]*StatelessNode

	// List of values, nil if this is an internal node.
	values map[byte][]byte

	stem []byte

	// node depth in the tree, in bits
	depth byte

	// child count, used for the special case in
	// commitment calculations.
	count uint

	// Cache the field representation of the hash
	// of the current node.
	hash *Fr

	// Cache the commitment value
	commitment, c1, c2 *Point

	committer Committer
}

func NewStateless() *StatelessNode {
	cfg, _ := GetConfig()
	return &StatelessNode{
		children:   make(map[byte]*StatelessNode),
		hash:       new(Fr).SetZero(),
		committer:  cfg,
		commitment: Generator(),
	}
}

func NewStatelessWithCommitment(point *Point) *StatelessNode {
	var (
		xfr Fr
	)
	toFr(&xfr, point)
	cfg, _ := GetConfig()
	return &StatelessNode{
		children:   make(map[byte]*StatelessNode),
		hash:       &xfr,
		committer:  cfg,
		commitment: point,
	}
}

func (n *StatelessNode) Children() []VerkleNode {
	var children [256]VerkleNode
	for i := range children {
		if n.children[byte(i)] != nil {
			children[i] = n.children[byte(i)]
		} else {
			children[i] = Empty(struct{}{})
		}
	}
	return children[:]
}

func (n *StatelessNode) SetChild(i int, v VerkleNode) error {
	if i >= NodeWidth {
		return errors.New("child index higher than node width")
	}
	c, ok := v.(*StatelessNode)
	if !ok {
		return errors.New("inserting non-stateless node into a stateless node")
	}
	n.children[byte(i)] = c
	return nil
}

// SetStem is an accessor for a stateless leaf node stem
func (n *StatelessNode) SetStem(stem []byte) {
	n.stem = stem
}

func (n *StatelessNode) Insert(key []byte, value []byte, resolver NodeResolverFn) error {
	// if this is a leaf value and the stems are different, intermediate
	// nodes need to be inserted.
	if n.values != nil {
		// Need to add a new branch node to differentiate
		// between two keys, if the keys are different.
		// Otherwise, just update the key.
		if equalPaths(n.stem, key) {
			if n.values[key[31]] == nil {
				// only increase the count if no value is
				// overwritten.
				n.count++
			}
			n.values[key[31]] = value
			// TODO: instead of invalidating the commitment
			// and recalulating it entirely, compute the diff.
			n.hash = nil
			n.ComputeCommitment()
		} else {
			// A new branch node has to be inserted. Depending
			// on the next word in both keys, a recursion into
			// the moved leaf node can occur.
			nextWordInExistingKey := offset2key(n.stem, n.depth)
			oldExtNode := &StatelessNode{
				depth:      n.depth + 1,
				committer:  n.committer,
				count:      n.count,
				values:     n.values,
				stem:       n.stem,
				commitment: new(Point),
				hash:       n.hash,
				c1:         n.c1,
				c2:         n.c2,
			}
			n.children = map[byte]*StatelessNode{
				nextWordInExistingKey: oldExtNode,
			}
			n.values = nil
			n.stem = nil
			n.c1 = nil
			n.c2 = nil
			n.count++
			CopyPoint(oldExtNode.commitment, n.commitment)
			n.hash = new(Fr)

			nextWordInInsertedKey := offset2key(key, n.depth)
			if nextWordInInsertedKey != nextWordInExistingKey {
				// Next word differs, so the branching point
				// has been reached. Create the "new" child.
				n.children[nextWordInInsertedKey] = &StatelessNode{
					depth:     n.depth + 1,
					stem:      key[:31],
					values:    map[byte][]byte{key[31]: value},
					committer: n.committer,
					count:     1,
				}
				n.children[nextWordInInsertedKey].ComputeCommitment()
			}

			// recurse into the newly created child
			if err := n.children[nextWordInInsertedKey].Insert(key, value, resolver); err != nil {
				return err
			}

			var poly [NodeWidth]Fr
			CopyFr(&poly[nextWordInExistingKey], oldExtNode.hash)
			if nextWordInExistingKey != nextWordInInsertedKey {
				CopyFr(&poly[nextWordInInsertedKey], n.children[nextWordInInsertedKey].hash)
			}
			n.commitment = n.committer.CommitToPoly(poly[:], NodeWidth-2)
			toFr(n.hash, n.commitment)
		}
	} else {
		// internal node
		nChild := offset2key(key, n.depth)

		// special case: missing child, insert a leaf
		cfg, _ := GetConfig()
		if n.children[nChild] == nil {
			n.children[nChild] = &StatelessNode{
				depth:      n.depth + 1,
				count:      1,
				values:     map[byte][]byte{key[31]: value},
				committer:  n.committer,
				stem:       key[:31],
				commitment: Generator(),
			}
			n.children[nChild].ComputeCommitment()
			var diff Point
			diff.ScalarMul(&cfg.conf.SRSPrecompPoints.SRS[nChild], n.children[nChild].hash)
			n.commitment.Add(n.commitment, &diff)
			toFr(n.hash, n.commitment)
			return nil
		}

		// Save the value of the initial child commitment
		var pre Fr
		CopyFr(&pre, n.children[nChild].hash)

		if err := n.children[nChild].Insert(key, value, resolver); err != nil {
			return err
		}

		// update the commitment
		var diff Point
		diff.ScalarMul(&cfg.conf.SRSPrecompPoints.SRS[nChild], pre.Sub(n.children[nChild].hash, &pre))
		n.commitment.Add(n.commitment, &diff)
	}

	toFr(n.hash, n.commitment)
	return nil
}

// insertStem is a translation of insert_node in the block explorer.
// It inserts a given stem in the tree, placing it as described
// by stemInfo. Its third parameters is the list of commitments
// that have not been assigned a node. It returns the same list,
// save the commitments that have been assigned a node during the
// call.
func (n *StatelessNode) insertStem(path []byte, stemInfo stemInfo, comms []*Point) ([]*Point, error) {
	if len(path) == 0 {
		return comms, errors.New("invalid path")
	}

	// path is 1 byte long, the leaf node must be created
	if len(path) == 1 {
		switch stemInfo.stemType & 3 {
		case extStatusAbsentEmpty:
			// nothing to do
		case extStatusAbsentOther:
			// insert poa stem
		case extStatusPresent:
			// insert stem
			n.children[path[0]] = NewStatelessWithCommitment(comms[0])
			comms = comms[1:]
			if stemInfo.has_c1 {
				n.children[path[0]].c1 = comms[0]
				comms = comms[1:]
			}
			if stemInfo.has_c2 {
				n.children[path[0]].c2 = comms[0]
				comms = comms[1:]
			}
			n.children[path[0]].values = stemInfo.values
			n.children[path[0]].stem = stemInfo.stem
			n.children[path[0]].depth = n.depth + 1
		}
		return comms, nil
	}

	// create the child node if missing
	if n.children[path[0]] == nil {
		n.children[path[0]] = NewStatelessWithCommitment(comms[0])
		comms = comms[1:]
		n.children[path[0]].depth = n.depth + 1
	}

	// recurse
	return n.children[path[0]].insertStem(path[1:], stemInfo, comms)
}

func (n *StatelessNode) insertValue(key, value []byte) error {
	// reached a leaf node ?
	if len(n.children) == 0 {
		if !bytes.Equal(key[:31], n.stem) {
			return errInsertIntoOtherStem
		}
		n.values[key[31]] = value
	} else { // no, recurse
		nChild := offset2key(key, n.depth)
		n.children[nChild].insertValue(key, value)
	}

	return nil
}

func (*StatelessNode) InsertOrdered([]byte, []byte, NodeFlushFn) error {
	return errNotSupportedInStateless
}

func (n *StatelessNode) Delete(key []byte, resolver NodeResolverFn) error {
	// Case of an ext node
	if n.values != nil {
		var zero [32]byte
		// Set the value to 0, data can not be deleted
		n.values[key[31]] = zero[:]
		n.hash = nil
		n.ComputeCommitment()
		return nil
	}

	nChild := offset2key(key, n.depth)
	child := n.children[nChild]
	var pre Fr
	CopyFr(&pre, child.hash)
	if err := child.Delete(key, resolver); err != nil {
		return err
	}

	pre.Sub(child.hash, &pre)

	var tmp Point
	cfg, _ := GetConfig()
	tmp.ScalarMul(&cfg.conf.SRSPrecompPoints.SRS[nChild], &pre)
	n.commitment.Add(n.commitment, &tmp)
	toFr(n.hash, n.commitment)
	return nil
}

func (n *StatelessNode) Get(k []byte, getter NodeResolverFn) ([]byte, error) {
	if n.values != nil {
		// if the stems are different, then the key is missing
		if bytes.Equal(n.stem, k[:31]) {
			return n.values[k[31]], nil
		}
		return nil, nil
	}

	nChild := offset2key(k, n.depth)

	child := n.children[nChild]
	if child == nil {
		return nil, nil
	}
	return child.Get(k, getter)
}

func (n *StatelessNode) ComputeCommitment() *Point {
	if n.hash != nil {
		return n.commitment
	}

	if n.count == 0 {
		if n.depth != 0 {
			panic("internal node should be empty node")
		}

		n.commitment = new(Point)
		n.commitment.Identity()
		n.hash = new(Fr)
		toFr(n.hash, n.commitment)
		return n.commitment
	}

	n.hash = new(Fr)

	if n.values == nil {
		// panic because this case should not be necessary
		panic("ComputeCommitment can not be called on a non-root value in stateless mode")
	}

	// leaf node: go over each value, and set them in the
	// polynomial for the corresponding suffix node.
	count1, count2 := 0, 0
	var poly, c1poly, c2poly [256]Fr
	poly[0].SetUint64(1)
	StemFromBytes(&poly[1], n.stem)

	for idx, val := range n.values {
		if idx < 128 {
			leafToComms(c1poly[idx<<1:], val)
			count1++
		} else {
			leafToComms(c2poly[(idx<<1)&0xFF:], val)
			count2++
		}
	}
	n.c1 = n.committer.CommitToPoly(c1poly[:], 256-count1)
	toFr(&poly[2], n.c1)
	n.c2 = n.committer.CommitToPoly(c2poly[:], 256-count2)
	toFr(&poly[3], n.c2)

	n.commitment = n.committer.CommitToPoly(poly[:], 252)
	toFr(n.hash, n.commitment)

	return n.commitment
}

func (n *StatelessNode) GetProofItems(keys keylist) (*ProofElements, []byte, [][]byte) {
	var (
		pe = &ProofElements{
			Cis:    []*Point{},
			Zis:    []byte{},
			Yis:    []*Fr{},
			ByPath: map[string]*Point{},
		}

		esses []byte   = nil // list of extension statuses
		poass [][]byte       // list of proof-of-absence stems
	)

	if len(n.values) == 0 {
		var (
			groups = groupKeys(keys, n.depth)
		)

		for _, group := range groups {
			childIdx := offset2key(group[0], n.depth)

			var yi Fr
			// when proving that a key is not in the tree
			if n.children[childIdx] == nil {
				yi.SetZero()
			} else {
				toFr(&yi, n.children[childIdx].commitment)
			}

			pe.Cis = append(pe.Cis, n.commitment)
			pe.Zis = append(pe.Zis, childIdx)
			pe.Yis = append(pe.Yis, &yi)
			pe.ByPath[string(group[0][:n.depth])] = n.commitment

		}

		// Loop over again, collecting the children's proof elements
		// This is because the order is breadth-first.
		for _, group := range groups {
			childIdx := offset2key(group[0], n.depth)

			// Special case of a proof of absence: no children
			// commitment, as the value is 0.
			if n.children[childIdx] == nil {
				// A question arises here: what if this proof of absence
				// corresponds to several stems? Should the ext status be
				// repeated as many times? It would be wasteful, so the
				// decoding code has to be aware of this corner case.
				esses = append(esses, extStatusAbsentEmpty|((n.depth+1)<<3))
				continue
			}

			pec, es, other := n.children[childIdx].GetProofItems(group)
			pe.Merge(pec)
			poass = append(poass, other...)
			esses = append(esses, es...)
		}
	} else {
		pe.Cis = append(pe.Cis, n.commitment, n.commitment)
		pe.Zis = append(pe.Zis, 0, 1)
		pe.Yis = append(pe.Yis, new(Fr).SetOne(), new(Fr).SetZero())
		StemFromBytes(pe.Yis[len(pe.Yis)-1], n.stem)

		// First pass: add top-level elements first
		var hasC1, hasC2 bool
		for _, key := range keys {
			hasC1 = hasC1 || (key[31] < 128)
			hasC2 = hasC2 || (key[31] >= 128)
			if hasC2 {
				break
			}
		}
		if hasC1 {
			var yi Fr
			toFr(&yi, n.c1)
			pe.Cis = append(pe.Cis, n.commitment)
			pe.Zis = append(pe.Zis, 2)
			pe.Yis = append(pe.Yis, &yi)
		}
		if hasC2 {
			var yi Fr
			toFr(&yi, n.c2)
			pe.Cis = append(pe.Cis, n.commitment)
			pe.Zis = append(pe.Zis, 3)
			pe.Yis = append(pe.Yis, &yi)
		}

		for _, key := range keys {
			pe.ByPath[string(key[:n.depth])] = n.commitment

			// Proof of absence: case of a differing stem.
			// Add an unopened stem-level node.
			if !equalPaths(n.stem, key) {
				// Corner case: don't add the poa stem if it's
				// already present as a proof-of-absence for a
				// different key, or for the same key (case of
				// multiple missing keys being absent).
				// The list of extension statuses has to be of
				// length 1 at this level, so skip otherwise.
				if len(esses) == 0 {
					esses = append(esses, extStatusAbsentOther|(n.depth<<3))
					poass = append(poass, n.stem)
				}
				continue
			}

			// corner case (see previous corner case): if a proof-of-absence
			// stem was found, and it now turns out the same stem is used as
			// a proof of presence, clear the proof-of-absence list to avoid
			// redundancy.
			if len(poass) > 0 {
				poass = nil
				esses = nil
			}

			var (
				suffix   = key[31]
				suffSlot = 2 + suffix/128 // slot in suffix tree
				scomm    *Point
			)

			if suffix < 128 {
				scomm = n.c1
			} else {
				scomm = n.c2
			}

			// Proof of absence: case of a missing suffix tree.
			//
			// The suffix tree for this value is missing, i.e. all
			// values in the extension-and-suffix tree are grouped
			// in the other suffix tree (e.g. C2 if we are looking
			// at C1).
			if scomm == nil {
				esses = append(esses, extStatusAbsentEmpty|(n.depth<<3))
				continue
			}

			slotPath := string(key[:n.depth]) + string([]byte{suffSlot})

			// Proof of absence: case of a missing value.
			//
			// Suffix tree is present as a child of the extension,
			// but does not contain the requested suffix. This can
			// only happen when the leaf has never been written to
			// since after deletion the value would be set to zero
			// but still contain the leaf marker 2^128.
			if n.values[suffix] == nil {
				pe.Cis = append(pe.Cis, scomm, scomm)
				pe.Zis = append(pe.Zis, 2*suffix, 2*suffix+1)
				pe.Yis = append(pe.Yis, &FrZero, &FrZero)
				if len(esses) == 0 || esses[len(esses)-1] != extStatusPresent|(n.depth<<3) {
					esses = append(esses, extStatusPresent|(n.depth<<3))
				}
				pe.ByPath[slotPath] = scomm
				continue
			}

			// suffix tree is present and contains the key
			var leaves [2]Fr
			leafToComms(leaves[:], n.values[suffix])
			pe.Cis = append(pe.Cis, scomm, scomm)
			pe.Zis = append(pe.Zis, 2*suffix, 2*suffix+1)
			pe.Yis = append(pe.Yis, &leaves[0], &leaves[1])
			if len(esses) == 0 || esses[len(esses)-1] != extStatusPresent|(n.depth<<3) {
				esses = append(esses, extStatusPresent|(n.depth<<3))
			}
			pe.ByPath[slotPath] = scomm
		}

	}
	return pe, esses, poass
}

func (*StatelessNode) Serialize() ([]byte, error) {
	return nil, errNotSupportedInStateless
}

func (n *StatelessNode) Copy() VerkleNode {
	ret := &StatelessNode{
		commitment: new(Point),
		depth:      n.depth,
		committer:  n.committer,
		count:      n.count,
	}

	if n.children != nil {
		ret.children = make(map[byte]*StatelessNode, len(n.children))
		for i, child := range n.children {
			ret.children[i] = child.Copy().(*StatelessNode)
		}
	} else {
		ret.values = make(map[byte][]byte, len(n.values))
		for i, val := range n.values {
			ret.values[i] = make([]byte, len(val))
			copy(ret.values[i], val)
		}
	}

	if n.hash != nil {
		ret.hash = new(Fr)
		CopyFr(ret.hash, n.hash)
	}
	if n.commitment != nil {
		CopyPoint(ret.commitment, n.commitment)
	}

	return ret
}

func (n *StatelessNode) toDot(parent, path string) string {
	n.ComputeCommitment()
	me := fmt.Sprintf("internal%s", path)
	var ret string
	if len(n.values) != 0 {
		var c1bytes, c2bytes [32]byte
		if n.c1 != nil {
			c1bytes = n.c1.Bytes()
		}
		if n.c2 != nil {
			c2bytes = n.c2.Bytes()
		}
		ret = fmt.Sprintf("leaf%s [label=\"L: %x\nC: %x\nC₁: %x\nC₂:%x\"]\n%s -> leaf%s\n", path, n.hash.Bytes(), n.commitment.Bytes(), c1bytes, c2bytes, parent, path)
		for i, v := range n.values {
			if v != nil {
				ret = fmt.Sprintf("%sval%s%x [label=\"%x\"]\nleaf%s -> val%s%x\n", ret, path, i, v, path, path, i)
			}
		}
	} else {
		ret = fmt.Sprintf("%s [label=\"I: %x\"]\n", me, n.hash.BytesLE())
		if len(parent) > 0 {
			ret += fmt.Sprintf(" %s -> %s\n", parent, me)
		}

		for i, child := range n.children {
			ret += child.toDot(me, fmt.Sprintf("%s%02x", path, i)) + "\n"
		}
	}

	return ret
}

func (n *StatelessNode) setDepth(d byte) {
	n.depth = d
}
