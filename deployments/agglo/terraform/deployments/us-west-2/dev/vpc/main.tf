provider "aws" {
  region = "us-west-2"
  profile = "default"
}

module "vpc" {
  source = "../../../../modules/vpc"

  vpc_cidr = "10.0.0.0/16"
  vpc_name = "kevin"
}
