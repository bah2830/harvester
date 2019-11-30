function init() {
    $.each(['harvest', 'copy', 'refresh', 'info', 'settings'], function (i, id) {
        $('#' + id).click(function () {
            external.invoke(id);
        });
    });

    external.invoke('refresh');
}

function renderTimers() {
    var rows = "";
    $.each(timers.data.tasks, function (i, timer) {
        rows += getTimerRow(timer);
    });
    $('#main-content').html(rows);
}

function getTimerRow(timer) {
    var icon = '<img src="/img/icons/jira.png" height="20px">';
    if (timer.harvest != undefined) {
        icon = '<img src="/img/icons/harvest.png" height="20px">';
    }

    var description = "";
    if (timer.jira != undefined) {
        description = timer.jira.fields.summary;
    } else if (timer.harvest != undefined) {
        description = timer.harvest.project.name;
    }
    if (description == undefined) {
        description = "";
    }

    var buttonValue = '<img src="/img/icons/play.png" height="20px">';
    var buttonOnClick = "startTimer('" + timer.key + "')";
    if (timer.running) {
        buttonValue = timer.runtime;
        buttonOnClick = "stopTimer('" + timer.key + "')";
    }

    return '<div class="d-flex flex-row align-middle task-timer"> \
        <div class="p-1">' + icon + '</div> \
        <div class="col text-truncate"> \
            <a href="javascript:openLink(\'' + timer.key + '\');" class="jira-link">' + timer.key + ': ' + description + '</a> \
        </div> \
        <div class="p-1"> \
            <button type="button" onclick="' + buttonOnClick + '" class="btn btn-dark btn-sm"> \
                ' + buttonValue + ' \
            </button > \
        </div > \
    </div>\n'
}

function timerUpdate() {
    renderTimers();
}

function startTimer(key) {
    external.invoke("start|" + key);
    return false;
}

function stopTimer(key) {
    external.invoke("stop|" + key);
    return false;
}

function openLink(key) {
    external.invoke("open|" + key);
    return false;
}