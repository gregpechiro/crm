$(document).ready(function() {
    $('span.calander').click(function(){
        $(this).parent().find('input').focus()
    });
});
