# Redirector App
This is a very simple openresty/nginx based HTTP redirector.

Configuration is defined in redis using a Hash where keys are the domain names and values are the location
they will be redirected to. Only defined domains will have certificates generated using Let's Encrypt.
This is managed through [Terraform](https://github.com/CruGlobal/cru-terraform/tree/master/applications/redirector/prod).
