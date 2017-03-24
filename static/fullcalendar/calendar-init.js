
/**
* Theme: Velonic Admin Template
* Author: Coderthemes
* Full calendar page
*/

!function($) {
    "use strict";

    var CalendarPage = function() {};

    CalendarPage.prototype.init = function() {

        //checking if plugin is available
        if ($.isFunction($.fn.fullCalendar)) {
            /* initialize the external events */
            $('#external-events .fc-event').each(function() {
                // create an Event Object (http://arshaw.com/fullcalendar/docs/event_data/Event_Object/)
                // it doesn't need to have a start or end
                var eventObject = {
                    title: $.trim($(this).text()) // use the element's text as the event title
                };

                // store the Event Object in the DOM element so we can get to it later
                $(this).data('eventObject', eventObject);

                // make the event draggable using jQuery UI
                $(this).draggable({
                    zIndex: 999,
                    revert: true, // will cause the event to go back to its
                    revertDuration: 0 //  original position after the drag
                });

            });

            /* initialize the calendar */
            var date = new Date();
            var d = date.getDate();
            var m = date.getMonth();
            var y = date.getFullYear();

            $('#calendar').fullCalendar({
                header: {
                    left: 'prev,next today',
                    center: 'title',
                    right: 'month,basicWeek,basicDay'
                },
                editable: true,
                eventLimit: true, // allow "more" link when too many events
                droppable: true, // this allows things to be dropped onto the calendar !!!
                drop: function(date, allDay) { // this function is called when something is dropped
		   var eventObject = $(this).data('eventObject');
		   eventObject.start = date;
		   eventObject.allDay = allDay;
		   $('#calendar').fullCalendar('renderEvent', eventObject, true);
                },
		eventSources: ['/calendar/events','/calendar']
	    });

             /*Add new event*/
            $("#add_event_form").on('submit', function(ev) {
                ev.preventDefault();
		var event_id = new Date().getTime();
		var event_name = $(this).find('#event_name').val();
		var event_date = $(this).find('#event_date').val();
		var eventObject = {
			id : event_id,
			title: event_name,
			start: event_date,
			allDay: true
		};
		// add new event to calendar / render new event
		$('#calendar').fullCalendar('renderEvent', eventObject, true);
		$.ajax({
			url: '/calendar/event',
			method: 'POST',
			data: eventObject,
			success: function(d) {
				if (d.err) {
					//alert('code:' + d.code + ', ' + 'msg: ' + d.msg);
                    $.Notification.autoHideNotify('error', 'top right', d.msg);
                    return
                }
                $.Notification.autoHideNotify('success', 'top right', d.msg);
			},
			error: function() {
				alert('some client side issue most likely...');
			}
		});
		// reset form
		this.reset();
            });
        } else {
            alert("Calendar plugin is not installed");
        }
    },
    //init
    $.CalendarPage = new CalendarPage, $.CalendarPage.Constructor = CalendarPage
}(window.jQuery),

//initializing
function($) {
    "use strict";
    $.CalendarPage.init()
}(window.jQuery);
