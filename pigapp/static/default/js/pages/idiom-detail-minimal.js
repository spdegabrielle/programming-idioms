function elem(tag, clazz, html) {
    // TODO: exists something standard for this?
    var element = document.createElement(tag);
    if(clazz)
        element.className = clazz;
    if(html)
        element.innerHTML = html;
    return element;
}

function elemText(tag, clazz, text) {
    // TODO: exists something standard for this?
    var element = document.createElement(tag);
    if(clazz)
        element.className = clazz;
    if(text)
        element.textContent = text;
    return element;
}

function emphasize(raw){
    // Emphasize the "underscored" identifier
    //
    // _x -> <span class="variable">x</span>
    //
    var refined = raw.replace( /\b_([\w$]*)/gm, "<span class=\"variable\">$1</span>");
    refined = refined.replace(/\n/g,"<br/>");
    return refined;
}

function renderImpl(impl) {
    var implNode = elem("div", "implementation");
    implNode.id = "impl-" + impl.Id;
    var lg = elem("h2", "", impl.LanguageName);
    implNode.appendChild(lg);

    if(impl.ImportsBlock){
        var piimports = elem("div", "piimports");
        var pre = elemText("pre", "", impl.ImportsBlock);
        piimports.appendChild(pre);
        implNode.appendChild(piimports);
    }

    var picode = elem("div", "picode");
    var pre = elemText("pre", "", impl.CodeBlock);
    picode.appendChild(pre);
    implNode.appendChild(picode);

    var comment = elem("div", "comment", emphasize(impl.AuthorComment));
    implNode.appendChild(comment);

    var links = elem("div", "external-links");
    var ul = elem("ul");
    if(impl.DemoURL) {
        var li = elem("li");
        var a = elem("a", "", "Demo 🗗");
        a.href = impl.DemoURL;
        a.target="_blank";
        a.rel="nofollow";
        li.appendChild(a);
        ul.appendChild(li);
    }
    if(impl.DocumentationURL) {
        var li = elem("li");
        var a = elem("a", "", "Doc 🗗");
        a.href = impl.DocumentationURL;
        a.target="_blank";
        a.rel="nofollow";
        li.appendChild(a);
        ul.appendChild(li);
    }
    if(impl.OriginalAttributionURL) {
        var li = elem("li");
        var a = elem("a", "", "Origin 🗗");
        a.href = impl.OriginalAttributionURL;
        a.target="_blank";
        a.rel="nofollow";
        li.appendChild(a);
        ul.appendChild(li);
    }
    links.appendChild(ul);
    implNode.appendChild(links);

    var impls = document.querySelector(".implementations");
    impls.appendChild(implNode);
    // console.log( implNode.id + " added!" );
}

function renderHeader() {
    var hh = document.getElementsByTagName("header");
    if(!hh.length) {
        console.error("Couldn't find header element");
        return;
    }
    var h = hh[0];
    while(h.firstChild) {
        h.removeChild(h.firstChild);
    }

    var hadd = function(code){
        h.insertAdjacentHTML('beforeend', code);
    }
    hadd('<a href="/"><img src="/default_20171211_/img/wheel_48x48.png" width="48" height="48" class="header_picto" /></a>');
    hadd('<h1><a href="/">Programming-Idioms</a></h1>');
    hadd('<a href="/random-idiom"><img src="/default_20171211_/img/dice_32x32.png" width="32" height="32" class="picto die" title="Go to a random idiom" /></a>');
    hadd('<form class="form-search" action="/search"> \
            <input type="text" class="search-query" placeholder="Keywords..." name="q" value="" required="required"> \
            <button type="submit">Search</button> \
          </form>');
}

function renderFooter() {
    // TODO
}

// Server-side rendering already includes the HTML for only
// a few impls.
// populateOtherImpls does client-side rendering of all other
// impls.
function populateOtherImpls(idiom) {
    idiom.Implementations.forEach(function(impl) {
        var nodeId = "impl-" + impl.Id;
        if( document.getElementById(nodeId) ){
            // console.log("Skipping existing " + nodeId);
        }else{
            renderImpl(impl);
        }
    });
}

function decorateImpls(idiom) {
    idiom.Implementations.forEach(function(impl) {
        var nodeId = "impl-" + impl.Id;
        var implNode = document.getElementById(nodeId);
        if( !implNode ){
            console.error("Couldn't find " + nodeId);
            return;
        }
        implNode.insertAdjacentHTML('beforeend',
            '<a href="/impl-edit/' 
            + idiom.Id + '/' 
            + impl.Id + '" class="edit hide-on-mobile" title="Edit this implementation">Edit</a>');
    });
}

function decorateSummary(idiom) {
    var nodes = document.getElementsByClassName("summary-large");
    if (!nodes || !nodes.item(0))
        return;
    nodes.item(0).insertAdjacentHTML('beforeend', 
        '<a href="/idiom-edit/' + 
        idiom.Id + 
        '" title="Edit the idiom statement" class="edit hide-on-mobile">Edit</a>');
}

//
// Execution!
//

renderHeader();

if(idiomPromise){
    idiomPromise
        .then(function(response) {
            // console.log("Got response");
            return response.json();
        })
        .then(function(idiom) {
            console.log("Got JSON of idiom " + idiom.Id);
            populateOtherImpls(idiom);
            decorateImpls(idiom);
            decorateSummary(idiom);
        });
}

renderFooter();