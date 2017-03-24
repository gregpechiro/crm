$('#setCompNotify').click(function() {
    $.ajax({
        url: '/cns/company/global/notify',
        type: 'GET',
        complete: function(resp) {
            var msg = '';
            if (resp.statusText === 'OK' && !resp.responseJSON.error) {
                msg = resp.responseJSON.msg;
            } else {
                msg = 'Unable to retrieve the last time customer service notifications were set'
            }
            swal({
                title: 'Company Service Notifications',
                text: msg + '.\nAre your sure your would like to set all customer service notifications for this month?',
                type: 'warning',
                showCancelButton: true,
                confirmButtonColor: '#33b86c',
                confirmButtonText: 'Yes',
                cancelButtonText: 'No',
                closeOnConfirm: true,
                closeOnCancel: true
            }, function(isConfirm) {
                if (isConfirm) {
                    $.ajax({
                        url: '/cns/company/global/notify',
                        type: 'POST',
                        complete: function(resp) {
                            var msg = '';
                            var type = '';
                            if (resp.statusText === 'OK' && !resp.responseJSON.error) {
                                msg = resp.responseJSON.msg;
                                type = 'success';
                            } else {
                                msg = 'Error setting customer service notifications'
                                type = 'error';
                            }
                            $.Notification.autoHideNotify(type, 'top right', msg);
                        }
                    });
                }
            });
        }
    });
});

$('#resetCompNotify').click(function() {
    $.ajax({
        url: '/cns/company/global/notify/reset',
        type: 'GET',
        complete: function(resp) {
            msg = '';
            if (resp.statusText === 'OK' && !resp.responseJSON.error) {
                msg = resp.responseJSON.msg;
            } else {
                msg = 'Unable to retrieve the last time customer service notifications were set'
            }
            swal({
                title: 'Company Service Notifications',
                text: msg + '.\nAre your sure your would like to reset all customer service notifications?',
                type: 'warning',
                showCancelButton: true,
                confirmButtonColor: '#33b86c',
                confirmButtonText: 'Yes',
                cancelButtonText: 'No',
                closeOnConfirm: true,
                closeOnCancel: true
            }, function(isConfirm) {
                if (isConfirm) {
                    $.ajax({
                        url: '/cns/company/global/notify/reset',
                        type: 'POST',
                        complete: function(resp) {
                            var msg = '';
                            var type = '';
                            if (resp.statusText === 'OK' && !resp.responseJSON.error) {
                                msg = resp.responseJSON.msg;
                                type = 'success';
                            } else {
                                msg = 'Error resetting customer service notifications'
                                type = 'error';
                            }
                            $.Notification.autoHideNotify(type, 'top right', msg);
                        }
                    });
                }
            });
        }
    });
});

$('#setDriverNotify').click(function() {
    $.ajax({
        url: '/cns/driver/global/notify',
        type: 'GET',
        complete: function(resp) {
            var msg = '';
            if (resp.statusText === 'OK' && !resp.responseJSON.error) {
                msg = resp.responseJSON.msg;
            } else {
                msg = 'Unable to retrieve the last time driver form notifications were set'
            }
            swal({
                title: 'Driver Form Notifications',
                text: msg + '.\nAre your sure your would like to set all driver form notifications for this month?',
                type: 'warning',
                showCancelButton: true,
                confirmButtonColor: '#33b86c',
                confirmButtonText: 'Yes',
                cancelButtonText: 'No',
                closeOnConfirm: true,
                closeOnCancel: true
            }, function(isConfirm) {
                if (isConfirm) {
                    $.ajax({
                        url: '/cns/driver/global/notify',
                        type: 'POST',
                        complete: function(resp) {
                            var msg = '';
                            var type = '';
                            if (resp.statusText === 'OK' && !resp.responseJSON.error) {
                                msg = resp.responseJSON.msg;
                                type = 'success';
                            } else {
                                msg = 'Error setting driver form notifications'
                                type = 'error';
                            }
                            $.Notification.autoHideNotify(type, 'top right', msg);
                        }
                    });
                }
            });
        }
    });
});
