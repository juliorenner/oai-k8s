package controllers

import (
	"context"
	"fmt"

	oaiv1beta1 "github.com/juliorenner/oai-k8s/operator/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8sScheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("split controller unit tests", func() {
	instance := &oaiv1beta1.Split{}
	instance.Name = "test"
	instance.Namespace = "testnamespace"
	instance.Spec.CoreIP = "192.1.1.1"
	DescribeTable("getResourceName", func(split SplitPiece, expected string) {
		Expect(getResourceName(instance, split)).To(Equal(expected))
	},
		Entry("cu resource name", CU, "cu-test"),
		Entry("du resource name", DU, "du-test"),
		Entry("ru resource name", RU, "ru-test"),
	)

	DescribeTable("getSplitObjectKey", func(split SplitPiece, expectedNamespace, expectedName string) {
		objectKey := getSplitObjectKey(instance, split)
		Expect(objectKey.Namespace).To(Equal(expectedNamespace))
		Expect(objectKey.Name).To(Equal(expectedName))
	},
		Entry("cu object key", CU, "testnamespace", "cu-test"),
		Entry("du object key", DU, "testnamespace", "du-test"),
		Entry("ru object key", RU, "testnamespace", "ru-test"),
	)

	DescribeTable("getCUConfigMapContent", func(resources []runtime.Object, expectedContent string) {
		fakeClient := getFakeClient()

		createResources(resources, fakeClient)
		reconciler := &SplitReconciler{Client: fakeClient}

		cmContent, err := reconciler.getCUConfigMapContent(instance)
		Expect(err).To(BeNil())
		Expect(cmContent).To(Equal(expectedContent))
	},
		Entry("pods not yet created", []runtime.Object{}, "upfaddress: 192.1.1.1\nlocaladdress: \nsouthaddress: \n"),
		Entry("pods created", []runtime.Object{getSplitPod(instance, CU), getSplitPod(instance, DU)},
			"upfaddress: 192.1.1.1\nlocaladdress: 192.1.1.243\nsouthaddress: 192.1.1.244\n"),
	)

	DescribeTable("getDUConfigMapContent", func(resources []runtime.Object, expectedContent string) {
		fakeClient := getFakeClient()

		createResources(resources, fakeClient)
		reconciler := &SplitReconciler{Client: fakeClient}

		cmContent, err := reconciler.getDUConfigMapContent(instance)
		Expect(err).To(BeNil())
		Expect(cmContent).To(Equal(expectedContent))
	},
		Entry("pods not yet created", []runtime.Object{},
			"northaddress: \nlocaladdress: \nsouthaddress: \n"),
		Entry("pods created", []runtime.Object{
			getSplitPod(instance, CU),
			getSplitPod(instance, DU),
			getSplitPod(instance, RU),
		},
			"northaddress: 192.1.1.243\nlocaladdress: 192.1.1.244\nsouthaddress: 192.1.1.245\n"),
	)

	DescribeTable("getRUConfigMapContent", func(resources []runtime.Object, expectedContent string) {
		fakeClient := getFakeClient()

		createResources(resources, fakeClient)
		reconciler := &SplitReconciler{Client: fakeClient}

		cmContent, err := reconciler.getRUConfigMapContent(instance)
		Expect(err).To(BeNil())
		Expect(cmContent).To(Equal(expectedContent))
	},
		Entry("pods not yet created", []runtime.Object{}, "northaddress: \nlocaladdress: \n"),
		Entry("pods created", []runtime.Object{
			getSplitPod(instance, DU),
			getSplitPod(instance, RU),
		},
			"northaddress: 192.1.1.244\nlocaladdress: 192.1.1.245\n"),
	)

	DescribeTable("syncValuesConfigMap", func(resources []runtime.Object, isUpdate bool) {
		log := zap.New(zap.UseDevMode(true))
		scheme := runtime.NewScheme()
		Expect(k8sScheme.AddToScheme(scheme)).To(BeNil())
		Expect(oaiv1beta1.AddToScheme(scheme)).To(BeNil())
		fakeClient := fake.NewFakeClientWithScheme(scheme)

		reconciler := &SplitReconciler{Client: fakeClient, Scheme: scheme}

		createResources(resources, fakeClient)

		err := reconciler.syncValuesConfigMap(instance, log)
		Expect(err).To(BeNil())
		cuCM := &v1.ConfigMap{}
		Expect(fakeClient.Get(context.Background(), getSplitObjectKey(instance, CU), cuCM)).To(BeNil())
		Expect(cuCM.Data["values"]).To(Not(BeEmpty()))
		duCM := &v1.ConfigMap{}
		Expect(fakeClient.Get(context.Background(), getSplitObjectKey(instance, DU), duCM)).To(BeNil())
		Expect(duCM.Data["values"]).To(Not(BeEmpty()))
		ruCM := &v1.ConfigMap{}
		Expect(fakeClient.Get(context.Background(), getSplitObjectKey(instance, RU), ruCM)).To(BeNil())
		Expect(ruCM.Data["values"]).To(Not(BeEmpty()))
		if isUpdate {
			Expect(cuCM.Data["values"]).To(Not(Equal("dummy value")))
			Expect(duCM.Data["values"]).To(Not(Equal("dummy value")))
			Expect(ruCM.Data["values"]).To(Not(Equal("dummy value")))
		}
	},
		Entry("create config maps - pods not created", []runtime.Object{}, false),
		Entry("create config maps - pods created", []runtime.Object{
			getSplitPod(instance, CU),
			getSplitPod(instance, DU),
			getSplitPod(instance, RU),
		}, false),
		Entry("update config maps", []runtime.Object{
			getSplitPod(instance, CU),
			getSplitPod(instance, DU),
			getSplitPod(instance, RU),
			getSplitConfigMap(instance, CU),
			getSplitConfigMap(instance, DU),
			getSplitConfigMap(instance, RU),
		}, true),
	)

	DescribeTable("syncTemplatesConfigMap", func(shouldCreateTemplates, isUpdate, isErrorExpected bool) {
		log := zap.New(zap.UseDevMode(true))
		scheme := runtime.NewScheme()
		Expect(k8sScheme.AddToScheme(scheme)).To(BeNil())
		Expect(oaiv1beta1.AddToScheme(scheme)).To(BeNil())
		fakeClient := fake.NewFakeClientWithScheme(scheme)

		reconciler := &SplitReconciler{Client: fakeClient, Scheme: scheme}

		if shouldCreateTemplates {
			templates := getTemplateConfigMaps()
			resources := []runtime.Object{}
			for _, v := range templates {
				resources = append(resources, v)
			}
			createResources(resources, fakeClient)
		}

		if isUpdate {
			templates := getTemplateConfigMaps()
			resources := []runtime.Object{}
			for _, v := range templates {
				v.Namespace = instance.Namespace
				v.Data["template"] = "should_be_overwritten"
				resources = append(resources, v)
			}
			createResources(resources, fakeClient)
		}

		err := reconciler.syncTemplatesConfigMap(instance.Namespace, log)
		Expect(err != nil).To(Equal(isErrorExpected))

		if !isErrorExpected {
			objKey := getCreatedTemplatesObjectKeys(instance.Namespace)
			for _, value := range objKey {
				cm := &v1.ConfigMap{}
				Expect(fakeClient.Get(context.Background(), value, cm)).To(BeNil())
				Expect(cm.Data["template"]).To(Equal("dummy"), fmt.Sprintf("cm %s data do not match", cm.Name))
			}
		}
	},
		Entry("error - config maps in operator namespace not created", false, false, true),
		Entry("sync - create", true, false, false),
		Entry("sync - update", true, true, false),
	)
})

func getSplitPod(instance *oaiv1beta1.Split, split SplitPiece) *v1.Pod {
	var ip string
	switch split {
	case CU:
		ip = "192.1.1.243"
	case DU:
		ip = "192.1.1.244"
	case RU:
		ip = "192.1.1.245"
	}

	objectKey := getSplitObjectKey(instance, split)
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectKey.Name,
			Namespace: objectKey.Namespace,
			Labels: map[string]string{
				"split":       string(split),
				"split-owner": instance.Name,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "dummy",
					Image: "dummy",
				},
			},
		},
		Status: v1.PodStatus{
			PodIP: ip,
		},
	}
}

func getSplitConfigMap(instance *oaiv1beta1.Split, split SplitPiece) *v1.ConfigMap {
	objKey := getSplitObjectKey(instance, split)
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: objKey.Namespace,
			Name:      objKey.Name,
		},
		Data: map[string]string{
			"template": "dummy value",
		},
	}
}

func getTemplateConfigMaps() []*v1.ConfigMap {
	templates := []*v1.ConfigMap{}
	for _, templateName := range TemplateConfigMaps {
		templates = append(templates, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      templateName,
				Namespace: operatorNamespace,
			},
			Data: map[string]string{
				"template": "dummy",
			},
		})
	}
	return templates
}

func getCreatedTemplatesObjectKeys(namespace string) []types.NamespacedName {
	objKeys := []types.NamespacedName{}
	for _, templateName := range TemplateConfigMaps {
		objKeys = append(objKeys, types.NamespacedName{
			Namespace: namespace,
			Name:      templateName,
		})
	}

	return objKeys
}

func getFakeClient() client.Client {
	scheme := runtime.NewScheme()
	Expect(k8sScheme.AddToScheme(scheme)).To(BeNil())
	return fake.NewFakeClientWithScheme(scheme)
}

func createResources(resources []runtime.Object, k8sClient client.Client) {
	for _, resource := range resources {
		Expect(k8sClient.Create(context.Background(), resource)).To(BeNil())
	}
}
