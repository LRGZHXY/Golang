package practice

func swapPairs(head *ListNode) *ListNode {
	dummy := &ListNode{}
	dummy.Next = head
	pre := dummy
	for pre.Next != nil && pre.Next.Next != nil {
		pre.Next = head.Next
		next := head.Next.Next
		head.Next.Next = head
		head.Next = next
		pre = pre.Next.Next
	}
	return dummy.Next
}
