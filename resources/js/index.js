let index = {
    init: function () {
        document.addEventListener('astilectron-ready', function () {
            index.refresh();

            astilectron.onMessage(function(message) {
                if (message.Type == 'error') {
                    astilectron.showErrorBox(message.Type, message.Message);
                }
                if (message.Type == 'renderTimers') {
                    index.renderTimers(message.Timers);
                }
            });
        });

        $('#refresh').click(function () {
            index.refresh();
        });
    },
    renderTimers: function (timers) {
        var rows = "";
        $.each(timers, function(i, timer) {
            rows += index.getTimerRow(timer);
        })
        $('#timer-list').html(rows);
    },
    getTimerRow: function (timer) {
        console.log(timer);
        var icon = '<img src="/resources/img/icons/jira.png" height="20px">';
        if (timer.Harvest != undefined) {
            icon = '<img src="/resources/img/icons/harvest.png" height="20px">';
        }

        var description = '';
        if (timer.Jira != undefined) {
            description = timer.Jira.fields.description
        } else {
            description = timer.Harvest.project.name
        }

        return '<div class="row"> \
            <div class="col-11 text-truncate"> ' + icon + ' ' + timer.Key + ': ' + description + '</div> \
            <div class="col"> </div> \
        </div>\n'
    },
    refresh: function () {
        astilectron.sendMessage("refresh");
    },
};