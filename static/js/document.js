function readyData() {
	var submitData = {
		redirect: redirect,
		id: id,
		data: JSON.stringify(inputTools.getJSON())
	};
	return submitData;
}

function send(url) {
	$.ajax({
		url: url,
		data: readyData(),
		method: 'POST',
		success: function(resp) {
			test = resp
			if (resp.status === 'success') {
				window.location.pathname = resp.redirect
			}
		},
		error: function(data) {
			console.log(data);
		}
	});
}

$(document).ready(function() {

	$('button#save').click(function() {
		send(url + '/save');
	});

	$('button#complete').click(function() {
		$('div#invalidMsg').addClass('hide');
		if (inputTools.validate()) {
			send(url + '/complete');
		} else {
			$('div#invalidMsg').removeClass('hide');
			$('html, body').animate({ scrollTop: 0 }, 'fast');
		}
	});

	// if (data !== '') {
	// 	inputTools.fill(JSON.parse(data));
	// }

	if (!$.isEmptyObject(data)) {
		inputTools.fill(data);
	}
});
