package controllers

//var _ = Describe("split controller unit tests", func() {
//	DescribeTable("validateRUPlacement", func(memory, cpu int64, isErrorExpected bool) {
//		topology := &oaiv1beta1.Topology{
//			Nodes: map[string]*oaiv1beta1.Node{
//				"node1": {
//					Interfaces: nil,
//					Core:       false,
//					Resources: oaiv1beta1.Resources{
//						MemoryAvailable: NewMemoryQuantity(memory),
//						CPUAvailable:    NewCPUQuantity(cpu),
//					},
//				},
//				"node2": {
//					Interfaces: nil,
//					Core:       false,
//					Resources: oaiv1beta1.Resources{
//						MemoryAvailable: NewMemoryQuantity(2048),
//						CPUAvailable:    NewCPUQuantity(3000),
//					},
//				},
//			},
//		}
//
//		splitPlacer := &oaiv1beta1.SplitsPlacer{Spec: oaiv1beta1.SplitsPlacerSpec{RUs: []oaiv1beta1.RUPosition{
//			{
//				SplitName: "split-1",
//				RUNode:    "node1",
//			},
//		}}}
//		log := zap.New(zap.UseDevMode(true))
//		reconciler := &SplitsPlacerReconciler{}
//		err := reconciler.validateRUPlacement(splitPlacer, topology, log)
//
//		Expect(err != nil).To(Equal(isErrorExpected))
//		Expect(topology.Nodes["node1"].Resources.MemoryAvailable.Value()).To(Equal(memory - SplitMemoryRequestValue))
//		Expect(topology.Nodes["node1"].Resources.CPUAvailable.Value()).To(Equal(cpu - SplitCPURequestValue))
//	},
//		Entry("available resources", int64(4098), int64(1000), false),
//		Entry("no resources available", int64(200), int64(200), true))
//})
