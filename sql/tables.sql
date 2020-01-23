USE team;

SET FOREIGN_KEY_CHECKS = 0;
 
DROP TABLE IF EXISTS teams;

DROP TABLE IF EXISTS projects;

DROP TABLE IF EXISTS members;

DROP TABLE IF EXISTS skills;

DROP TABLE IF EXISTS languages;

SET FOREIGN_KEY_CHECKS = 1;

CREATE TABLE teams (
    id int not null PRIMARY key auto_increment,
    leader varchar(255) not null,
    team_name varchar(25) not null,
    open_roles int not null,
    size int not null,
    last_active int
);

CREATE TABLE members (
    id int not null PRIMARY key auto_increment,
    user_id int not null,
    member_email varchar(255) not null,
    member_role varchar(40) not null,
    team_id int,
    FOREIGN KEY(team_id) REFERENCES teams(id)
);

CREATE TABLE skills (
    id int not null PRIMARY key auto_increment,
    skill_name varchar(100) not null,
    team_id int,
    FOREIGN KEY(team_id) REFERENCES teams(id)
);

CREATE TABLE projects (
    id int not null PRIMARY key auto_increment,
    goal varchar(1200) not null,
    project_name varchar(30) not null,
    github_link varchar(255) not null,
    complexity int not null,
    duration int not null,
    team_id int,
    FOREIGN KEY(team_id) REFERENCES teams(id)
);

CREATE TABLE languages (
    id int not null PRIMARY key auto_increment,
    lang_name varchar(100) not null,
    team_id int,
    FOREIGN KEY(team_id) REFERENCES teams(id)
);