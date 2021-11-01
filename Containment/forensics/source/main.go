package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"log"
	"net/url"
)

const (
	roleName = "forensics-role"
	policyName = "forensics-policy"
)

func main() {
	bucketArn := flag.String("bucketArn","", "Arn of the bucket, e.g. arn:aws:s3:::forensics/")
	namespace := flag.String("namespace", "forensics-system", "Service account namespace")
	serviceAccount := flag.String("serviceAccount", "forensics-worker", "Service account name")
	clusterName := flag.String("clusterName", "", "Name of cluster (required)")
	flag.Parse()
	flag.VisitAll(func (f *flag.Flag) {
		if f.Value.String()=="" {
			log.Fatalln(f.Name, "is not set")
		}
	})
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if cfg.Region == "" {
		log.Fatalln("Could not find region in default profile")
	}
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	iamPolicy := "{\n    \"Version\": \"2012-10-17\",\n    \"Statement\": [\n        {\n            \"Sid\": \"VisualEditor0\",\n            \"Effect\": \"Allow\",\n            \"Action\": [\n                \"s3:PutObject\",\n                \"s3:GetObject\"\n            ],\n            \"Resource\": \"" + *bucketArn + "*\"\n        }\n    ]\n}"

	k8s := eks.NewFromConfig(cfg)
	clusterOutput, err := k8s.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: clusterName,
	})
	if err != nil {
		log.Fatalln(err)
	}
	oidcEndpointUrl := clusterOutput.Cluster.Identity.Oidc.Issuer
	u, _ := url.Parse(*oidcEndpointUrl)
	oidcEndpoint := u.Host + u.Path
	log.Println(oidcEndpoint)

	svc := iam.NewFromConfig(cfg)
	oidcProviders, err := svc.ListOpenIDConnectProviders(ctx, &iam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		log.Println(err)
	}
	p := oidcProviders.OpenIDConnectProviderList
	oidcEndpointArn := p[0].Arn
	log.Println(*oidcEndpointArn)

	var policyArn string
	//create IAM policy
	cpo, err := svc.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyDocument: aws.String(iamPolicy),
		PolicyName:     aws.String(policyName),
		Description:    aws.String("allows writes to the specified s3 bucket"),
	})
	//if policy exists, get policy Arn
	if err != nil {
		var ne *types.EntityAlreadyExistsException
		if errors.As(err, &ne) {
			lpo, _ := svc.ListPolicies(ctx, &iam.ListPoliciesInput{Scope: "Local"})
			for _, v := range lpo.Policies {
				if aws.ToString(v.PolicyName) == "forensics-policy" {
					policyArn = aws.ToString(v.Arn)
				}
			}
		}
	} else {
		policyArn = aws.ToString(cpo.Policy.Arn)
	}

	trustPolicy := "{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Effect\": \"Allow\",\n      \"Principal\": {\n        \"Federated\": \"" + *oidcEndpointArn + "\"\n      },\n      \"Action\": \"sts:AssumeRoleWithWebIdentity\",\n      \"Condition\": {\n        \"StringEquals\": {\n          \"" + oidcEndpoint + ":sub\": \"system:serviceaccount:" + *namespace + ":" + *serviceAccount + "\"\n        }\n      }\n    }\n  ]\n}"
	//create IAM role
	cro, err := svc.CreateRole(ctx, &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(trustPolicy),
		RoleName:                 aws.String(roleName),
		Description:              aws.String("Role for writing forensic data to an s3 bucket"),
	})
	//if role already exists
	if err != nil {
		var ne *types.EntityAlreadyExistsException
		if errors.As(err, &ne) {
			log.Println("error:", ne)
			larp, err := svc.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{RoleName: aws.String("forensics-role")})
			if err != nil {
				log.Println(err)
			}
			//if the role doesn't have a policy, attach policyArn
			if len(larp.AttachedPolicies) == 0 {
				log.Println("No policies attached to role")
				svc.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
					RoleName: aws.String(roleName),
					PolicyArn: aws.String(policyArn),
				})
				gro, _ := svc.GetRole(ctx, &iam.GetRoleInput{RoleName: aws.String("forensics-role")})
				log.Println(aws.ToString(gro.Role.Arn))
				return
			//if the role has a policy, log the role Arn
			} else {
				gro, _ := svc.GetRole(ctx, &iam.GetRoleInput{RoleName: aws.String("forensics-role")})
				log.Println(aws.ToString(gro.Role.Arn))
				return
			}
		}
	// create role, and attach policy
	} else {
		_, err = svc.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: aws.String(policyArn),
		})
	}
	if err != nil {
		log.Println(err)
	}
	fmt.Println(aws.ToString(cro.Role.Arn))
}
