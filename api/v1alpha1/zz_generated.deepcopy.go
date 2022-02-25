//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Destinations) DeepCopyInto(out *Destinations) {
	*out = *in
	if in.RecentlyCalled != nil {
		in, out := &in.RecentlyCalled, &out.RecentlyCalled
		*out = new(Timestamp)
		(*in).DeepCopyInto(*out)
	}
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Destinations.
func (in *Destinations) DeepCopy() *Destinations {
	if in == nil {
		return nil
	}
	out := new(Destinations)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Fence) DeepCopyInto(out *Fence) {
	*out = *in
	if in.WormholePort != nil {
		in, out := &in.WormholePort, &out.WormholePort
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Fence.
func (in *Fence) DeepCopy() *Fence {
	if in == nil {
		return nil
	}
	out := new(Fence)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecyclingStrategy) DeepCopyInto(out *RecyclingStrategy) {
	*out = *in
	if in.Stable != nil {
		in, out := &in.Stable, &out.Stable
		*out = new(RecyclingStrategy_Stable)
		(*in).DeepCopyInto(*out)
	}
	if in.Deadline != nil {
		in, out := &in.Deadline, &out.Deadline
		*out = new(RecyclingStrategy_Deadline)
		(*in).DeepCopyInto(*out)
	}
	if in.Auto != nil {
		in, out := &in.Auto, &out.Auto
		*out = new(RecyclingStrategy_Auto)
		(*in).DeepCopyInto(*out)
	}
	if in.RecentlyCalled != nil {
		in, out := &in.RecentlyCalled, &out.RecentlyCalled
		*out = new(Timestamp)
		(*in).DeepCopyInto(*out)
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecyclingStrategy.
func (in *RecyclingStrategy) DeepCopy() *RecyclingStrategy {
	if in == nil {
		return nil
	}
	out := new(RecyclingStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecyclingStrategy_Auto) DeepCopyInto(out *RecyclingStrategy_Auto) {
	*out = *in
	if in.Duration != nil {
		in, out := &in.Duration, &out.Duration
		*out = new(Timestamp)
		(*in).DeepCopyInto(*out)
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecyclingStrategy_Auto.
func (in *RecyclingStrategy_Auto) DeepCopy() *RecyclingStrategy_Auto {
	if in == nil {
		return nil
	}
	out := new(RecyclingStrategy_Auto)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecyclingStrategy_Deadline) DeepCopyInto(out *RecyclingStrategy_Deadline) {
	*out = *in
	if in.Expire != nil {
		in, out := &in.Expire, &out.Expire
		*out = new(Timestamp)
		(*in).DeepCopyInto(*out)
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecyclingStrategy_Deadline.
func (in *RecyclingStrategy_Deadline) DeepCopy() *RecyclingStrategy_Deadline {
	if in == nil {
		return nil
	}
	out := new(RecyclingStrategy_Deadline)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecyclingStrategy_Stable) DeepCopyInto(out *RecyclingStrategy_Stable) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecyclingStrategy_Stable.
func (in *RecyclingStrategy_Stable) DeepCopy() *RecyclingStrategy_Stable {
	if in == nil {
		return nil
	}
	out := new(RecyclingStrategy_Stable)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Selector) DeepCopyInto(out *Selector) {
	*out = *in
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Selector.
func (in *Selector) DeepCopy() *Selector {
	if in == nil {
		return nil
	}
	out := new(Selector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceFence) DeepCopyInto(out *ServiceFence) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceFence.
func (in *ServiceFence) DeepCopy() *ServiceFence {
	if in == nil {
		return nil
	}
	out := new(ServiceFence)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceFence) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceFenceList) DeepCopyInto(out *ServiceFenceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceFence, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceFenceList.
func (in *ServiceFenceList) DeepCopy() *ServiceFenceList {
	if in == nil {
		return nil
	}
	out := new(ServiceFenceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceFenceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceFenceSpec) DeepCopyInto(out *ServiceFenceSpec) {
	*out = *in
	if in.Host != nil {
		in, out := &in.Host, &out.Host
		*out = make(map[string]*RecyclingStrategy, len(*in))
		for key, val := range *in {
			var outVal *RecyclingStrategy
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(RecyclingStrategy)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.NamespaceSelector != nil {
		in, out := &in.NamespaceSelector, &out.NamespaceSelector
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.LabelSelector != nil {
		in, out := &in.LabelSelector, &out.LabelSelector
		*out = make([]*Selector, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Selector)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.WorkloadSelector != nil {
		in, out := &in.WorkloadSelector, &out.WorkloadSelector
		*out = new(WorkloadSelector)
		(*in).DeepCopyInto(*out)
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceFenceSpec.
func (in *ServiceFenceSpec) DeepCopy() *ServiceFenceSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceFenceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceFenceStatus) DeepCopyInto(out *ServiceFenceStatus) {
	*out = *in
	if in.Domains != nil {
		in, out := &in.Domains, &out.Domains
		*out = make(map[string]*Destinations, len(*in))
		for key, val := range *in {
			var outVal *Destinations
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Destinations)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.MetricStatus != nil {
		in, out := &in.MetricStatus, &out.MetricStatus
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Visitor != nil {
		in, out := &in.Visitor, &out.Visitor
		*out = make(map[string]bool, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceFenceStatus.
func (in *ServiceFenceStatus) DeepCopy() *ServiceFenceStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceFenceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Timestamp) DeepCopyInto(out *Timestamp) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Timestamp.
func (in *Timestamp) DeepCopy() *Timestamp {
	if in == nil {
		return nil
	}
	out := new(Timestamp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkloadSelector) DeepCopyInto(out *WorkloadSelector) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkloadSelector.
func (in *WorkloadSelector) DeepCopy() *WorkloadSelector {
	if in == nil {
		return nil
	}
	out := new(WorkloadSelector)
	in.DeepCopyInto(out)
	return out
}
