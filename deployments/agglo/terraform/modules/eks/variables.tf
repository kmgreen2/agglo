variable "vpc_name" {
    type = string
}

variable "cluster_name" {
    type = string
}

variable "ec2_ssh_key" {
    type = string
    default = "Kevin's MBP"
}

variable "subnet_cidrs" {
    type = list(string)
}

variable "subnet_azs" {
    type = list(string)
}
