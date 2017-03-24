function Uploader() {
    this.init()
}

Uploader.prototype = {
    KB: 1024,
    MB: 1024 * 1024,
    fileTypes: ["image/jpeg", "image/png"],
    fileTypeErrorMsg: "Incorrect File type. Only JPEG and PNG files",
    defaultText: "Select File",
    maxSize: 2097152,
    maxSizeMsg: "File too large. Max size 2MB",
    updateFileInfo: function(input) {
        var value = input.val();
        var fileMatch = value.match(/([^\/\\]+)$/);
        var fileName = '';
        if (fileMatch == null) {
            if (input.attr('data-display') !== '' && input.attr('data-display') !== undefined) {
                fileName = input.attr('data-display');
            } else {
                fileName = this.defaultText;
            }
        } else {
            fileName = fileMatch[1];
        }
        $('label[for^="' + input.attr('id') + '"]').text(fileName);

        var form = input.parents('#uploader');

        var inputs = form.find('input.uploader');
        var allFiles = true;
        for (var i = 0; i < inputs.length; i++) {
            if (inputs[i].value == "") {
                allFiles = false;
            }
        }
        if (allFiles) {
            form.find('button#upload').removeAttr("disabled");
        } else {
            form.find('button#upload').attr("disabled", "disabled");
        }
    },
    fileCheck: function(input) {
        if ($('input[id="' + input.id + '"]')[0].files.length > 0) {
            var size = $('input[id="' + input.id + '"]')[0].files[0].size;
            var type = $('input[id="' + input.id + '"]')[0].files[0].type;
            if (size > this.maxSize) {
                $('input[id="' + input.id + '"]')[0].type = "text";
                $('input[id="' + input.id + '"]')[0].type = "file";
                this.displayError(this.maxSizeMsg)
                return
            }
        	if (this.fileTypes.indexOf(type) > -1) {
        		return;
        	} else {
                console.log(type);
        		$('input[id="' + input.id + '"]')[0].type = "text";
        		$('input[id="' + input.id + '"]')[0].type = "file";
                this.displayError(this.fileTypeErrorMsg);
        	}
        }
    },
    uploadClickAction: function(fn) {
        $('button[id="upload"]').click(function() {
            fn()
        });
    },
    displayError: function(msg) {
        alert(msg)
    },
    init: function() {
        var uploadButton = $("button#upload");
        var uploadInput = $("input.uploader");
        var uploadForm = $("form#uploader");
        var err = '';

        if (0 == uploadForm.length) {
            err += 'Upload form must have an id of "uploader"\n';
        }
        if (0 == uploadInput.length) {
            err += 'File input must have a class of "uploader"\n';
        }
        if (0 == uploadButton.length) {
            err += 'Submit button must have and id of "upload"\n';
        }

        if (err == '') {
            $("input.uploader").change(function() {
                uploader.fileCheck(this);
                uploader.updateFileInfo($(this));
            });
        } else {
            console.log(err);
        }
    }
}

var uploader = new Uploader;
