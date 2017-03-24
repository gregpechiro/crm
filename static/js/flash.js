
//setCookie('flash', baseEnc('alertError:Error adding job to cart. Please try again in a few minutes'), 7);

function baseEnc(s) {
    s = btoa(s);
    while (s[s.length - 1] === '=' ) {
        s =  s.slice(0, s.length - 1)
    }
    return s;
}

function setCookie(cname, cvalue, exdays) {
    var d = new Date();
    d.setTime(d.getTime() + (exdays*24*60*60*1000));
    var expires = "expires="+d.toUTCString();
    document.cookie = cname + '=' + cvalue + '; ' + expires +'; path=/';
}

function setSuccessFlash(msg) {
    setCookie('flash', baseEnc('alertSuccess:' + msg), 7);
}

function setErrorFlash(msg) {
    setCookie('flash', baseEnc('alertError:' + msg), 7);
}
