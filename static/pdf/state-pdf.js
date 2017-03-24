
var pages = $('.page');
var pdfData;
// var doc = new jsPDF('p', 'in');
var doc;
try {
    doc = new jsPDF('p', 'px', [918,1188]);
} catch(e) {
    console.log(e);
}
var w = 918;
var h = 1188;
var name;

if (pages.length > 0 && $(pages[0]).width() > 950) {
    var doc = new jsPDF('l', 'px', [918,1188]);
    w = 1188;
    h = 918;
}

$(document).ready(function() {
    $('button#archive').click(function() {
        swal({
            title: 'Archiving Form',
            text: 'Please wait while we archive your form.\n Do not leave or refresh this page.',
            type: 'warning',
            showConfirmButton: false
        });
        name = $('input#archive-name').val();
        if (name === '') {
            setErrorFlash('Error archiving form. No name was specified.<br>Please enter a name and try again.');
            window.location.reload();
            return
        }
        convertPages();
    });
});

function convertPages() {
    $('body').append($('<div class="white-out"></div>'));
    getCanvas(0);
}
function getCanvasHelper(page) {
    page.find('input, textarea').addClass('no-border');

    window.scrollTo(0, 0);

    var srcEl = page[0];
    var scaleFactor = 2;

    var originalWidth = srcEl.offsetWidth;
    var originalHeight = srcEl.offsetHeight;
    // Force px size (no %, EMs, etc)
    srcEl.style.width = originalWidth + "px";
    srcEl.style.height = originalHeight + "px";

    // Position the element at the top left of the document because of bugs in html2canvas. The bug exists when supplying a custom canvas, and offsets the rendering on the custom canvas based on the offset of the source element on the page; thus the source element MUST be at 0, 0.
    // See html2canvas issues #790, #820, #893, #922
    srcEl.style.position = "absolute";
    srcEl.style.top = "0";
    srcEl.style.left = "0";

    // Create scaled canvas
    var scaledCanvas = document.createElement("canvas");
    scaledCanvas.width = originalWidth * scaleFactor;
    scaledCanvas.height = originalHeight * scaleFactor;
    scaledCanvas.style.width = originalWidth + "px";
    scaledCanvas.style.height = originalHeight + "px";
    var scaledContext = scaledCanvas.getContext("2d");
    scaledContext.scale(scaleFactor, scaleFactor);

    return html2canvas(srcEl, { canvas: scaledCanvas })
}

function getCanvas(idx) {
    getCanvasHelper($(pages[idx])).then(function(canvas) {
        $(pages[idx]).find('input, texterea').removeClass('no-border');

        var img = canvas.toDataURL('image/png');
        try {
            doc.addImage(img, 'JPEG', 0, 0, w, h, idx, 'FAST');
        } catch(e) {
            console.log(e);
            return
        }

        if (idx === (pages.length - 1)) {
            uploadDoc();
            console.log('done');
            return
        }

        doc.addPage();
        getCanvas(idx + 1);
    });
}

function uploadDoc() {
    console.log(1);
    var data = new FormData();
console.log(2);
    data.append('file', doc.output('blob'));
console.log(3);
    name = $('input#archive-name').val();
console.log(4);
    if (name === '') {
        setErrorFlash('Error archiving form. No name was specified.<br>Please enter a name and try again.');
        window.location.reload();
        return
    }
    console.log(5);
    data.append('name', name);
console.log(6);
    $.ajax({
        url: '/cns/company/' + companyId + '/archive',
        method: 'POST',
        data: data,
        processData: false,
        contentType: false,
        success: function(resp) {
            if (resp.error) {
                setErrorFlash(resp.msg);
                window.location.reload();
                return;
            }
            setSuccessFlash(resp.msg);
            window.location = '/cns/company/' + companyId + '/file';
            return;
        },
        error: function(d, s) {
            setErrorFlash('Erro archiving form as ' + name + '.pdf. Please try again.');
            window.location.reload();
            return;
        }
    });
}
