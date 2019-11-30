function init() {
    $('#jiraUrl').val(settings.data.jira.url);
    $('#jiraUser').val(settings.data.jira.user);
    $('#jiraPass').val(settings.data.jira.pass);
    $('#harvestUser').val(settings.data.harvest.user);
    $('#harvestPass').val(settings.data.harvest.pass);

    $('#save').click(function () {
        save();
    });
}

function save() {
    var settings = {
        jira: {
            url: $('#jiraUrl').val(),
            user: $('#jiraUser').val(),
            pass: $('#jiraPass').val()
        },
        harvest: {
            user: $('#harvestUser').val(),
            pass: $('#harvestPass').val()
        }
    }
    external.invoke(JSON.stringify(settings));
}