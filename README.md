CLI Stack Changer Plugin
=====================
This plugin can be used to view and update applications running on the outdated Lucid64 backend due to Canonical discontinuing support at the end of April 2015.

##Installation

#####Install from Repo (v.6.10.0+)
  ```
  $ cf add-plugin-repo CF-Community http://plugins.cloudfoundry.org/
  $ cf install-plugin "Stack Changer" -r CF-Community
  ```
#####Install from Url (v.6.8.0+)
OSX
  ```
  cf install-plugin https://github.com/simonleung8/cli-stack-changer/raw/master/bin/osx/cli-stack-changer
  ```

linux64:
  ```
  cf install-plugin https://github.com/simonleung8/cli-stack-changer/raw/master/bin/linux64/cli-stack-changer
  ```

windows64:
  ```
  cf install-plugin https://github.com/simonleung8/cli-stack-changer/raw/master/bin/win64/cli-stack-changer
  ```


#####Install from Binary file (v.6.7.0)


- Download the binary [`win64`](https://github.com/simonleung8/cli-stack-changer/raw/master/bin/win64/cli-stack-changer) [`linux64`](https://github.com/simonleung8/cli-stack-changer/raw/master/bin/linux64/cli-stack-changer) [`osx`](https://github.com/simonleung8/cli-stack-changer/raw/master/bin/osx/cli-stack-changer)
- Install plugin `$ cf install-plugin <binary_name>`
  
##Full Command List

| command | usage | description|
| :--------------- |:---------------| :------------|
|`stack-change`| `cf stack-change App_Name` |Update stacks for apps from lucid64 to cflinuxfs2. Restart started apps.<br><br>Options:<br>`-o`: organization<br>`-s`: space<br>`-p`: # of concurrent threads in a batch|
|`stack-list`| `cf stack-list App_Name` |List all apps running on stack lucid64.<br><br>Options:<br>`-o`: organization<br>`-s`: space|



