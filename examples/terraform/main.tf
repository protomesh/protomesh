terraform {
  required_providers {
    protomesh = {
      source = "protomesh/protomesh"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }

  }
}

provider "random" {}

provider "protomesh" {
  address    = "localhost:6680"
  enable_tls = true
}

resource "random_uuid" "ping_pong_do_pong" {
}

resource "protomesh_aws_lambda_grpc" "ping_pong_do_pong" {

  namespace = "proxy"

  resource_id = random_uuid.ping_pong_do_pong.result
  name        = "Route to PingPong.DoPong method"

  node {
    full_method_name = "/protomesh.services.v1.PingPongService/DoPing"
    function_name    = "protomeshPingback"
    qualifier        = "$LATEST"
  }

}
