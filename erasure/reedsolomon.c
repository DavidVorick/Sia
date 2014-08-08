// This file is intended to be used exclusively with reedsolomon.go, and relies
// on the error checking and usage patterns of reedsolomon.go
#include "longhair/include/cauchy_256.h"
#include <stdlib.h>
#include <string.h>

// encodeRedundancy takes as input a 'k', the number of nonredundant segments
// and an 'm', the number of redundant segments. k + m must be less than 255.
// bytesPerSegment is the size of each segment, and then originalBlock is a pointer
// to the original data, which is assumed to be of size k * bytesPerSegment
//
// The return value is a block of data m * bytesPerSegment which contains all of
// the redundant data. The data does not get segmented into pieces in this
// function.
static char *encodeRedundancy(int k, int m, int bytesPerSegment, char *originalBlock) {
	// Verify that correct library is linked.
	if (cauchy_256_init()) {
		return NULL;
	}

	// Break original data into segments using pointer math.
	const unsigned char *originalSegments[k];
	int i;
	for (i = 0; i < k; i++) {
		originalSegments[i] = (const unsigned char*)&originalBlock[i * bytesPerSegment];
	}

	// allocate space for redundant segments
	char *redundantSegments = calloc(sizeof(unsigned char), m * bytesPerSegment);

	// encode the redundant segments using longhair
	if (cauchy_256_encode(k, m, originalSegments, redundantSegments, bytesPerSegment)) {
		return NULL;
	}

	return redundantSegments;
}

// recoverData takes as input 'k', the number of nonredundant segments and 'm',
// the number of redundant segments. 'bytesPerSegment' indicates how large each
// segment is. remainingSegments is a pointer to a block of data that contains
// exactly 'k' uncorrupted segments. 'remainingSegmentIndices' indicate which
// segments of the original set the uncorrupted ones correspond with.
//
// The data is edited and sorted in place. Upon returning, 'remainingSegments'
// will be the original data in order.
static int recover(int k, int m, int bytesPerSegment, unsigned char *remainingSegments, unsigned char *remainingSegmentIndices) {
	// Verify that the longhair library is linked.
	if (cauchy_256_init()) {
		return 1;
	}

	// copy remainingSegments into its own data, which results in much cleaner
	// code during the sorting phase.
	unsigned char *workingMemory = calloc(k+m, bytesPerSegment);
	memcpy(workingMemory, remainingSegments, k*bytesPerSegment);

	// Longhair has a block data structure, which is composed of a pointer to
	// data, and a row. It recovers the file into the blocks, but doesn't do any
	// sorting, adjusting the 'row' field of the block to indiciate how the
	// sorting should happen. 'remainingSegments' contains the input 'k' pieces,
	// and will also contain the unsorted ouput pieces after calling
	// caughy_256_decode.
	Block blocks[k];
	unsigned char i, j;
	for (i = 0; i < k; i++) {
		blocks[i].data = &workingMemory[i * bytesPerSegment];
		blocks[i].row = remainingSegmentIndices[i];
	}
	
	// decode redundant segments into original segments
	if (cauchy_256_decode(k, m, blocks, bytesPerSegment)) {
		return 2;
	}

	// Perform a collapseSort, which creates an array of len 'k + m' and puts
	// each index in its corresponding location. This sorts the data, but leaves
	// m gaps. Then the array is iterated through and 'collapsed', removing the
	// gaps. This puts the blocks into order and copies the memory into
	// recovered.
	int *ordering = calloc(k+m, sizeof(int));
	for (i = 0; i < k; i++) {
		ordering[blocks[i].row] = (int)i + 1; // little hack to avoid initializing the array to -1
	}
	j = 0;
	for (i = 0; i < k+m; i++) {
		if (ordering[i]) {
			memcpy(&remainingSegments[j * bytesPerSegment], blocks[ordering[i]-1].data, bytesPerSegment);
			j++;
		}
	}

    return 0;
}
