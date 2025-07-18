package practice

func getIntersectionNode(headA, headB *ListNode) *ListNode {
	lenA, lenB := 0, 0
	var fast, slow *ListNode
	a, b := headA, headB
	for a != nil {
		a = a.Next
		lenA++
	}
	for b != nil {
		b = b.Next
		lenB++
	}
	dif := 0
	if lenA > lenB {
		dif = lenA - lenB
		fast, slow = headA, headB
	} else {
		dif = lenB - lenA
		fast, slow = headB, headA
	}
	for dif > 0 {
		fast = fast.Next
		dif--
	}
	for fast != slow {
		fast = fast.Next
		slow = slow.Next
	}
	return fast
}
