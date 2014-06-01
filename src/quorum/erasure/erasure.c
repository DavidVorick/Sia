#include "longhair/include/cauchy_256.h"
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

// encodeRedundancy takes as input a 'k', the number of nonredundant segments
// and an 'm', the number of redundant segments. k + m must be less than 256.
// bytesPerSegment is the size of each segment, and then originalBlock is a pointer
// to the original data, which is assumed to be of size k * bytesPerSegment
//
// The return value is a block of data m * bytesPerSegment which contains all of
// the redundant data. The data does not get segmented into pieces in this
// function.
//
// encodeRedundancy does not do any error checking, all of that must happen
// in the calling function.


static char *encodeRedundancy(int k, int m, int bytesPerSegment, char *originalBlock) {
	// verify that correct library is linked
	if(cauchy_256_init()) {
		fprintf(stderr, "Failed to init cauchy_256 library");
		exit(1);
	}

	// break original data into segments
	const unsigned char *originalSegments[k];
	int i;
	for(i = 0; i < k; i++) {
		originalSegments[i] = (const unsigned char*)&originalBlock[i * bytesPerSegment];
	}

	// allocate space for redundant segments
	char *redundantSegments = calloc(sizeof(unsigned char), m * bytesPerSegment);

	// encode the redundant segments
	if(cauchy_256_encode(k, m, originalSegments, redundantSegments, bytesPerSegment)) {
		fprintf(stderr, "Failed to encode pieces - strange...");
		exit(1);
	}

	return redundantSegments;
}

static void inPlaceSort(Block segmentInfo[], int bytesPerSegment, int start, int end);

static void swap(Block segmentInfo[], int bytesPerSegment, int i, int j) {
	// temporary data holders
	char tempData[bytesPerSegment];
	char tempIndex = segmentInfo[i].row;
	memcpy(tempData, segmentInfo[i].data, bytesPerSegment);
	
	// storing j's info into i
	segmentInfo[i].row = segmentInfo[j].row;
	memcpy(segmentInfo[i].data, segmentInfo[j].data, bytesPerSegment);

	// storing i's old info into j via tmep
	segmentInfo[j].row = tempIndex;
	memcpy(segmentInfo[j].data, tempData, bytesPerSegment);
}

static void workingMerge(Block segmentInfo[], int bytesPerSegment, int left_start, int left_end, int right_start, int right_end, int work_area) {
	// perform swap for both halves of data block
	while (left_start < left_end && right_start < right_end) {
		swap(segmentInfo, bytesPerSegment, work_area++ ,segmentInfo[left_start].row < segmentInfo[right_start].row ? left_start++: right_start++);
	}
	// if there are more lower elements than higher elements (remaining) process those as well
	while (left_start < left_end) {
		swap(segmentInfo, bytesPerSegment, work_area++, left_start++);
	}
	// if there are more higher elements than lower elemtns (remaingin) process those as well
	while (right_start < right_end) {
		swap(segmentInfo, bytesPerSegment, work_area++, right_start++);
	}
	
}

static void workingSort(Block segmentInfo[], int bytesPerSegment, int start, int end, int work_area) {
	int middle;
	// check the given set of data is more than one element
	if (end - start > 1) {
		middle = start + (end - start)/2;
		inPlaceSort(segmentInfo, bytesPerSegment, start, middle);
		inPlaceSort(segmentInfo, bytesPerSegment, middle, end);
		workingMerge(segmentInfo, bytesPerSegment, start, middle, middle, end, work_area);
	} else {
	// other wise just run the swap method
		while (start < end) {
			swap(segmentInfo, bytesPerSegment, start++, work_area++);
		}
	}
}

static void inPlaceSort(Block segmentInfo[], int bytesPerSegment, int start, int end) {
	int m, n, w;
	// check to see of the given data block is more than one element
	if (end - start > 1) {
		m = start + (end - start)/2;
		w = start + end - m;
		workingSort(segmentInfo, bytesPerSegment, start, m, w);

		// check to see if the block of data is greater than 2 elements
		while (w - start > 2) {
			n = w;
			w = start + (n - start + 1)/2;
			workingSort(segmentInfo, bytesPerSegment, w, n, start);
			workingMerge(segmentInfo, bytesPerSegment, start, start+n-w, n, end, w);
		}
		// shift over to insertion sort to finish up
		for (n = w; n > start; --n) {
			for (m = n; m < end && segmentInfo[m].row < segmentInfo[m-1].row; ++m) {
				swap(segmentInfo, bytesPerSegment, m, m - 1);

			}
		}
	}
}

// recoverData takes as input 'k', the number of nonredundant segments and 'm',
// the number of redundant segments. 'bytesPerSegment' indicates how large each
// segment is. remainingSegments is a pointer to a block of data that contains
// exactly 'k' uncorrupted segments. 'remainingSegmentIndicies' indicate which
// segments of the original set the uncorrupted ones correspond with.
//
// The data is edited and sorted in place. Upon returning, 'remainingSegments'
// will be the original data in order.
static void recoverData(int k, int m, int bytesPerSegment, unsigned char *remainingSegments, unsigned char *remainingSegmentIndicies) {
	// verify that correct library is linked
	if(cauchy_256_init()) {
		//error
		exit(1);
	}

	// create blocks around data using the indicies
	Block segmentInfo[k];
	int i, j;
	for(i = 0; i < k; i++) {
		segmentInfo[i].data = &remainingSegments[i * bytesPerSegment];
		segmentInfo[i].row = remainingSegmentIndicies[i];
	}
	
	// decode redundant segments into original segments
	if(cauchy_256_decode(k, m, segmentInfo, bytesPerSegment)) {
		//error
		exit(1);
	}

	/* sort memory back into original order */
	/* because I want to push sooner rather than later, I'm using an n^2 sort */
	/* eventually, I'll implement something more efficient */

	// allocate space to copy memory into
	
	inPlaceSort(segmentInfo, bytesPerSegment, 0, k);
}


