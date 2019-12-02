class App extends React.Component {
  render() {
    return React.createElement("div", null, React.createElement(Toolbar, null), React.createElement(Timers, null));
  }

}

if (typeof timers !== 'undefined') {
  const render = () => ReactDOM.render(React.createElement(App, null), document.getElementById('app'));

  timers.render = render;
}
class Settings extends React.Component {
  submit(e) {
    e.preventDefault();
  }

  save() {
    var settings = {
      jira: {
        url: document.getElementById('jiraUrl').value,
        user: document.getElementById('jiraUser').value,
        pass: document.getElementById('jiraPass').value
      },
      harvest: {
        user: document.getElementById('harvestUser').value,
        pass: document.getElementById('harvestPass').value
      }
    };
    external.invoke(JSON.stringify(settings));
  }

  render() {
    return React.createElement("form", {
      onSubmit: this.submit
    }, React.createElement("h4", null, "Jira"), React.createElement("div", {
      className: "form-group row"
    }, React.createElement("label", {
      htmlFor: "jiraUrl",
      className: "col-4 col-form-label"
    }, "URL"), React.createElement("div", {
      className: "col"
    }, React.createElement("input", {
      type: "text",
      className: "form-control",
      id: "jiraUrl",
      placeholder: "url",
      defaultValue: settings.data.jira.url
    }))), React.createElement("div", {
      className: "form-group row"
    }, React.createElement("label", {
      htmlFor: "jiraUser",
      className: "col-4 col-form-label"
    }, "Username"), React.createElement("div", {
      className: "col"
    }, React.createElement("input", {
      type: "text",
      className: "form-control",
      id: "jiraUser",
      placeholder: "username",
      defaultValue: settings.data.jira.user
    }))), React.createElement("div", {
      className: "form-group row"
    }, React.createElement("label", {
      htmlFor: "jiraPass",
      className: "col-4 col-form-label"
    }, "Password"), React.createElement("div", {
      className: "col"
    }, React.createElement("input", {
      type: "password",
      className: "form-control",
      id: "jiraPass",
      placeholder: "password",
      defaultValue: settings.data.jira.pass
    }))), React.createElement("br", null), React.createElement("br", null), React.createElement("h4", null, "Harvest"), React.createElement("div", {
      className: "form-group row"
    }, React.createElement("label", {
      htmlFor: "harvestUser",
      className: "col-4 col-form-label"
    }, "Username"), React.createElement("div", {
      className: "col"
    }, React.createElement("input", {
      type: "text",
      className: "form-control",
      id: "harvestUser",
      placeholder: "username",
      defaultValue: settings.data.harvest.user
    }))), React.createElement("div", {
      className: "form-group row"
    }, React.createElement("label", {
      htmlFor: "harvestPass",
      className: "col-4 col-form-label"
    }, "Password"), React.createElement("div", {
      className: "col"
    }, React.createElement("input", {
      type: "password",
      className: "form-control",
      id: "harvestPass",
      placeholder: "password",
      defaultValue: settings.data.harvest.pass
    }))), React.createElement("br", null), React.createElement("button", {
      id: "save",
      className: "btn btn-primary btn-block",
      onClick: this.save
    }, "Save"));
  }

}

if (typeof settings !== 'undefined') {
  const render = () => ReactDOM.render(React.createElement(Settings, null), document.getElementById('app'));

  settings.render = render;
  render();
}
class Timer extends React.Component {
  constructor(props) {
    super(props);
    this.stopTimer = this.stopTimer.bind(this);
    this.startTimer = this.startTimer.bind(this);
    this.openLink = this.openLink.bind(this);
  }

  stopTimer() {
    external.invoke("stop|" + this.props.timer.key);
  }

  startTimer() {
    external.invoke("start|" + this.props.timer.key);
  }

  openLink() {
    external.invoke("start|" + this.props.timer.key);
  }

  render() {
    const timer = this.props.timer;
    let iconSrc = '/img/icons/jira.png';

    if (timer.harvest != undefined) {
      iconSrc = '/img/icons/harvest.png';
    }

    const icon = React.createElement("img", {
      src: iconSrc,
      height: "20px"
    });
    let description = "";

    if (timer.jira != undefined) {
      description = timer.jira.fields.summary;
    } else if (timer.harvest != undefined) {
      description = timer.harvest.project.name;
    }

    const playImg = '/img/icons/play.png';
    const button = React.createElement("button", {
      type: "button",
      onClick: timer.running ? this.stopTimer : this.startTimer,
      className: "btn btn-dark btn-sm"
    }, timer.running ? timer.runtime : React.createElement("img", {
      src: playImg,
      height: "20px"
    }));
    return React.createElement("div", {
      className: "d-flex flex-row align-middle task-timer"
    }, React.createElement("div", {
      className: "p-1"
    }, icon), React.createElement("div", {
      className: "col text-truncate"
    }, React.createElement("a", {
      href: "#",
      onClick: this.openLink,
      className: "jira-link"
    }, timer.key, ": ", description)), React.createElement("div", {
      className: "p-1"
    }, button));
  }

}
class Timers extends React.Component {
  render() {
    var self = this;
    const rows = [];
    timers.data.tasks.forEach(function (timer, i) {
      rows.push(React.createElement(Timer, {
        key: i,
        timer: timer
      }));
    });
    return React.createElement("div", {
      id: "main-content"
    }, rows);
  }

}
class TimeSheet extends React.Component {
  render() {
    return React.createElement("div", null, "hi");
  }

}

if (typeof timesheet !== 'undefined') {
  const render = () => ReactDOM.render(React.createElement(TimeSheet, null), document.getElementById('app'));

  timesheet.render = render;
}
class Toolbar extends React.Component {
  refresh() {
    external.invoke('refresh');
  }

  settings() {
    external.invoke('settings');
  }

  timesheet() {
    external.invoke('timesheet');
  }

  render() {
    return React.createElement("nav", {
      className: "navbar navbar-expand-lg fixed-top navbar-dark d-flex flex-row"
    }, React.createElement("div", {
      className: "p-2"
    }, React.createElement("img", {
      id: "copy",
      src: "/img/icons/copy.png",
      height: "20px"
    })), React.createElement("div", {
      className: "p-2"
    }, React.createElement("img", {
      id: "refresh",
      onClick: this.refresh,
      src: "/img/icons/refresh.png",
      height: "20px"
    })), React.createElement("div", {
      className: "col"
    }, "\xA0"), React.createElement("div", {
      className: "p-2"
    }, React.createElement("img", {
      id: "timesheet",
      onClick: this.timesheet,
      src: "/img/icons/info.png",
      height: "20px"
    })), React.createElement("div", {
      className: "p-2"
    }, React.createElement("img", {
      id: "settings",
      onClick: this.settings,
      src: "/img/icons/settings.png",
      height: "20px"
    })));
  }

}
