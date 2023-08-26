# kubeconfig-solo

Store kubeconfigs individually under `~/.kube/configs/<env>/<context-name>.yaml` to help working with multiple clusters simultaneously and in an isolated fashion.

Use [kshell function](https://github.com/arunvelsriram/dotfiles/blob/45791c31dabc6cf9772080bbf501f103bdaf3ef3/oh-my-zsh-custom/plugins/kubectl/kubectl.plugin.zsh#L20) to switch context with fuzzy search.


## Usage

```shell
make compile
./out/kubeconfig-solo /path/to/clusters.yaml

kshell [env]
```

```
Usage of ./out/kubeconfig-solo:
  -c string
        create kubeconfigs for the given cluster name only
  -e string
        create kubeconfigs for clusters belonging to given env
```
