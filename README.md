# trusty
static web shelf

```
cd
curl https://storage.googleapis.com/appengine-sdks/featured/go_appengine_sdk_linux_amd64-1.9.25.zip -o goappeng.zip
mv go_appengine go_appengine_old
unzip goappeng.zip 
more go_appengine/VERSION
export PATH=~/go_appengine:$PATH
cd ~/pumpkin/src/github.com/monopole
rm ~/.appcfg_oauth2_tokens 
~/go_appengine/goapp serve trusty
# vi trusty/app.yaml # application: lyrical-gantry-618
~/go_appengine/goapp deploy trusty
# then visit https://lyrical-gantry-618.appspot.com/
# or http://trustybike.net/

```

