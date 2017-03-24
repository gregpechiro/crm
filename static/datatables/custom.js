$(document).on('click', 'tr.clickable', function() {
    if (this.getAttribute('data-target') === '_blank') {
        window.open(this.getAttribute('data-url'));
        return
    }
    window.location.href = this.getAttribute('data-url');
});

$('#filter').on( 'keyup', function () {
    if (table !== undefined && !$.isEmptyObject(table)) {
        table.search( this.value ).draw();
    }
});

$('tr#search th').each(function () {
    var title = $(this).text();
    if (title != '') {
        $(this).html('<input id="columnSearch" data-column="' + title + '" type="text"/>');
    }
});

$('input#columnSearch').keyup(function() {
    var name = $(this).attr('data-column');
    var column = table.column(name + ':name');
    if (column.search() !== this.value ) {
        column.search(this.value).draw();
    }
});
