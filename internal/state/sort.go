package state

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func sortRoutes(routes []Route) {
	// stable sort is used so that the order of matches (as defined in each HTTPRoute rule) is preserved
	// this is important, because the winning match is the first match to win.
	sort.SliceStable(
		routes, func(i, j int) bool {
			return higherPriority(routes[i], routes[j])
		},
	)
}

/*
Returns true if route1 has a higher priority than route2.

From the spec:
 Precedence must be given to the Rule with the largest number of (Continuing on ties):
 - Characters in a matching non-wildcard hostname.
 - Characters in a matching hostname.
 - Characters in a matching path.
 - Header matches.
 - Query param matches.

 If ties still exist across multiple Routes, matching precedence MUST be determined in order of the following criteria, continuing on ties:
 - The oldest Route based on creation timestamp.
 - The Route appearing first in alphabetical order by “{namespace}/{name}”.

 If ties still exist within the Route that has been given precedence, matching precedence MUST be granted to the first matching rule meeting the above criteria.

higherPriority will determine precedence by comparing len(headers), len(query parameters), creation timestamp, and namespace name. The other criteria are handled by NGINX.
*/
func higherPriority(route1, route2 Route) bool {
	// Get the matches from the routes
	match1 := route1.GetMatch()
	match2 := route2.GetMatch()

	// If both matches exists then compare the number of header matches
	// The match with the largest number of header matches wins
	l1 := len(match1.Headers)
	l2 := len(match2.Headers)

	if l1 != l2 {
		return l1 > l2
	}
	// If the number of headers is equal then compare the number of query param matches
	// The match with the most query param matches wins
	l1 = len(match1.QueryParams)
	l2 = len(match2.QueryParams)

	if l1 != l2 {
		return l1 > l2
	}

	// If still tied, compare the object meta of the two routes.
	return lessObjectMeta(&route1.Source.ObjectMeta, &route2.Source.ObjectMeta)
}

func lessObjectMeta(meta1 *metav1.ObjectMeta, meta2 *metav1.ObjectMeta) bool {
	if meta1.CreationTimestamp.Equal(&meta2.CreationTimestamp) {
		return getResourceKey(meta1) < getResourceKey(meta2)
	}

	return meta1.CreationTimestamp.Before(&meta2.CreationTimestamp)
}

type mapper interface {
	Keys() []string
}

func getSortedKeys(m mapper) []string {
	keys := m.Keys()
	sort.Strings(keys)

	return keys
}
