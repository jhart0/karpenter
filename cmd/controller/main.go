/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/samber/lo"

	"github.com/aws/karpenter/pkg/cloudprovider"
	"github.com/aws/karpenter/pkg/controllers"
	"github.com/aws/karpenter/pkg/operator"
	"github.com/aws/karpenter/pkg/webhooks"

	"github.com/aws/karpenter-core/pkg/cloudprovider/metrics"
	corecontrollers "github.com/aws/karpenter-core/pkg/controllers"
	"github.com/aws/karpenter-core/pkg/controllers/state"
	coreoperator "github.com/aws/karpenter-core/pkg/operator"
	corewebhooks "github.com/aws/karpenter-core/pkg/webhooks"
)

func main() {
	ctx, op := operator.NewOperator(coreoperator.NewOperator())
	awsCloudProvider := cloudprovider.New(
		op.InstanceTypesProvider,
		op.InstanceProvider,
		op.GetClient(),
		op.AMIProvider,
	)
	lo.Must0(op.AddHealthzCheck("cloud-provider", awsCloudProvider.LivenessProbe))
	cloudProvider := metrics.Decorate(awsCloudProvider)

	op.
		WithControllers(ctx, corecontrollers.NewControllers(
			ctx,
			op.Clock,
			op.GetClient(),
			op.KubernetesInterface,
			state.NewCluster(op.Clock, op.GetClient(), cloudProvider),
			op.EventRecorder,
			cloudProvider,
		)...).
		WithWebhooks(corewebhooks.NewWebhooks()...).
		WithControllers(ctx, controllers.NewControllers(
			ctx,
			op.Session,
			op.Clock,
			op.GetClient(),
			op.EventRecorder,
			op.UnavailableOfferingsCache,
			op.SubnetProvider,
			op.SecurityGroupProvider,
			op.PricingProvider,
		)...).
		WithWebhooks(webhooks.NewWebhooks()...).
		Start(ctx)
}
